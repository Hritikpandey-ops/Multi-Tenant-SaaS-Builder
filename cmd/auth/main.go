package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"Desktop/multitenant/saas/internal/config"
	"Desktop/multitenant/saas/internal/database"
	"Desktop/multitenant/saas/internal/handlers"
	"Desktop/multitenant/saas/internal/jwt"
	"Desktop/multitenant/saas/internal/middleware"
	"Desktop/multitenant/saas/internal/repository"
)

func main() {
	// Load environment variables
	_ = godotenv.Load()

	// Initialize logger
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Set log level
	level, _ := zerolog.ParseLevel(cfg.App.LogLevel)
	zerolog.SetGlobalLevel(level)

	// Initialize database
	db, err := database.NewDatabase(cfg.Database, &log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := database.NewRedis(database.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}, &log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Initialize JWT manager
	jwtManager, err := jwt.NewJWTManager(cfg.JWT)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize JWT manager")
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.GORM)
	tenantRepo := repository.NewTenantRepository(db.GORM)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, tenantRepo, jwtManager, redisClient, &log.Logger)
	userHandler := handlers.NewUserHandler(userRepo, &log.Logger)
	tenantHandler := handlers.NewTenantHandler(tenantRepo, &log.Logger)

	// Set Gin mode
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Global middleware
	router.Use(middleware.Recovery(middleware.DefaultRecoveryConfig(&log.Logger)))
	router.Use(middleware.Logger(middleware.DefaultLoggerConfig(&log.Logger)))
	router.Use(middleware.CORS(middleware.DefaultCORSConfig(cfg.CORS.AllowedOrigins)))
	router.Use(middleware.SlidingWindowRateLimit(redisClient.Client, cfg.RateLimit.Requests, cfg.RateLimit.Window))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// Check database health
		if err := db.HealthCheck(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database unavailable",
			})
			return
		}

		// Check Redis health
		if err := redisClient.HealthCheck(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "redis unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "auth-service",
		})
	})

	// Metrics endpoint (for Prometheus)
	router.GET("/metrics", func(c *gin.Context) {
		// In production, integrate with Prometheus
		c.String(http.StatusOK, "# Metrics placeholder")
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			// Public routes (no authentication required)
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)

			// Protected routes (authentication required)
			protectedAuth := auth.Group("")
			protectedAuth.Use(middleware.AuthMiddleware(jwtManager))
			{
				protectedAuth.GET("/me", authHandler.GetCurrentUser)
				protectedAuth.POST("/logout", authHandler.Logout)
				protectedAuth.PUT("/me/password", authHandler.ChangePassword)
			}
		}

		// User routes
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(jwtManager))
		{
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.POST("", userHandler.CreateUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
			users.POST("/:id/invite", userHandler.InviteUser)
		}

		// Tenant routes
		tenant := v1.Group("/tenant")
		tenant.Use(middleware.AuthMiddleware(jwtManager))
		{
			tenant.GET("", tenantHandler.GetTenant)
			tenant.PUT("", tenantHandler.UpdateTenant)
			tenant.GET("/users", tenantHandler.GetTenantUsers)
		}
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Services.Auth.Host, cfg.Services.Auth.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Info().Str("addr", addr).Msg("Auth service starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start auth service")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Auth service shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Auth service exited")
}

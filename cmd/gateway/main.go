package main

import (
	"context"
	"fmt"
	"io"
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
	"Desktop/multitenant/saas/internal/jwt"
	"Desktop/multitenant/saas/internal/middleware"
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
			"service": "api-gateway",
		})
	})

	// Metrics endpoint (for Prometheus)
	router.GET("/metrics", func(c *gin.Context) {
		// In production, integrate with Prometheus
		c.String(http.StatusOK, "# Metrics placeholder")
	})

	// Service routes (proxy to backend services)
	api := router.Group("/api/v1")

	// Auth routes - forward to auth service (without /auth prefix to match backend)
	{
		// Public auth routes (no JWT required)
		api.POST("/auth/register", proxyToService(cfg.Services.Auth))
		api.POST("/auth/login", proxyToService(cfg.Services.Auth))
		api.POST("/auth/refresh", proxyToService(cfg.Services.Auth))

		// Protected auth routes (JWT required)
		protectedAuth := api.Use(middleware.AuthMiddleware(jwtManager))
		{
			protectedAuth.GET("/auth/me", proxyToService(cfg.Services.Auth))
			protectedAuth.POST("/auth/logout", proxyToService(cfg.Services.Auth))
			protectedAuth.PUT("/auth/me/password", proxyToService(cfg.Services.Auth))
		}
	}

	// User routes - all protected
	users := api.Group("/users").Use(middleware.AuthMiddleware(jwtManager))
	{
		users.GET("", proxyToService(cfg.Services.Auth))
		users.GET("/:id", proxyToService(cfg.Services.Auth))
		users.POST("", proxyToService(cfg.Services.Auth))
		users.PUT("/:id", proxyToService(cfg.Services.Auth))
		users.DELETE("/:id", proxyToService(cfg.Services.Auth))
		users.POST("/:id/invite", proxyToService(cfg.Services.Auth))
	}

	// Tenant routes - all protected
	tenant := api.Group("/tenant").Use(middleware.AuthMiddleware(jwtManager))
	{
		tenant.GET("", proxyToService(cfg.Services.Auth))
		tenant.PUT("", proxyToService(cfg.Services.Auth))
		tenant.GET("/users", proxyToService(cfg.Services.Auth))
		tenant.GET("/usage", proxyToService(cfg.Services.Auth))
	}

	// Billing routes
	billing := api.Group("/billing").Use(middleware.AuthMiddleware(jwtManager))
	{
		billing.GET("/subscription", proxyToService(cfg.Services.Billing))
		billing.POST("/subscription", proxyToService(cfg.Services.Billing))
		billing.PUT("/subscription", proxyToService(cfg.Services.Billing))
		billing.GET("/plans", proxyToService(cfg.Services.Billing))
		billing.POST("/webhook/stripe", proxyToService(cfg.Services.Billing))
	}

	// Analytics routes
	analytics := api.Group("/analytics").Use(middleware.AuthMiddleware(jwtManager))
	{
		analytics.POST("/events", proxyToService(cfg.Services.Analytics))
		analytics.GET("/events", proxyToService(cfg.Services.Analytics))
		analytics.GET("/dashboard", proxyToService(cfg.Services.Analytics))
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Services.Gateway.Host, cfg.Services.Gateway.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Info().Str("addr", addr).Msg("API Gateway starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start gateway")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("API Gateway shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("API Gateway exited")
}

// proxyToService creates a handler that proxies requests to backend services
func proxyToService(serviceCfg config.ServiceConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Build target URL
		targetURL := fmt.Sprintf("http://%s:%s%s?%s", serviceCfg.Host, serviceCfg.Port, c.Request.URL.Path, c.Request.URL.RawQuery)

		// Create new request
		proxyReq, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy request"})
			return
		}

		// Copy headers
		for key, values := range c.Request.Header {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}

		// Forward tenant context headers if present
		if tenantID, exists := c.Get("tenant_id"); exists {
			proxyReq.Header.Set("X-Tenant-ID", tenantID.(string))
		}
		if userID, exists := c.Get("user_id"); exists {
			proxyReq.Header.Set("X-User-ID", userID.(string))
		}

		// Execute request
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(proxyReq)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach backend service"})
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Copy response body
		body, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
	}
}

package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	App        AppConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	NATS       NATSConfig
	JWT        JWTConfig
	Services   ServicesConfig
	CORS       CORSConfig
	Stripe     StripeConfig
	RateLimit  RateLimitConfig
}

type AppConfig struct {
	Env     string
	Name    string
	LogLevel string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type NATSConfig struct {
	URL string
}

type JWTConfig struct {
	Secret               string
	ExpiresIn            time.Duration
	RefreshExpiresIn     time.Duration
	Issuer               string
	RSAPrivateKeyPath    string
	RSAPublicKeyPath     string
}

type ServicesConfig struct {
	Gateway   ServiceConfig
	Auth      ServiceConfig
	Billing   ServiceConfig
	Analytics ServiceConfig
	Tenant    ServiceConfig
}

type ServiceConfig struct {
	Host string
	Port string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type StripeConfig struct {
	APIKey         string
	WebhookSecret  string
}

type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists (ignore error in production)
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Env:     getEnv("APP_ENV", "development"),
			Name:    getEnv("APP_NAME", "multi-tenant-saas"),
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DATABASE_HOST", "localhost"),
			Port:     getEnv("DATABASE_PORT", "5432"),
			User:     getEnv("DATABASE_USER", "postgres"),
			Password: getEnv("DATABASE_PASSWORD", "postgres"),
			DBName:   getEnv("DATABASE_NAME", "multitenant"),
			SSLMode:  getEnv("DATABASE_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		NATS: NATSConfig{
			URL: getEnv("NATS_URL", "nats://localhost:4222"),
		},
		JWT: JWTConfig{
			Secret:            getEnv("JWT_SECRET", "change-me-in-production"),
			ExpiresIn:         getEnvAsDuration("JWT_EXPIRES_IN", 15*time.Minute),
			RefreshExpiresIn:  getEnvAsDuration("JWT_REFRESH_EXPIRES_IN", 7*24*time.Hour),
			Issuer:            getEnv("JWT_ISSUER", "multitenant-saas"),
			RSAPrivateKeyPath: getEnv("JWT_RSA_PRIVATE_KEY_PATH", "./keys/app.rsa"),
			RSAPublicKeyPath:  getEnv("JWT_RSA_PUBLIC_KEY_PATH", "./keys/app.rsa.pub"),
		},
		Services: ServicesConfig{
			Gateway: ServiceConfig{
				Host: getEnv("GATEWAY_HOST", "0.0.0.0"),
				Port: getEnv("GATEWAY_PORT", "3000"),
			},
			Auth: ServiceConfig{
				Host: getEnv("AUTH_SERVICE_HOST", "localhost"),
				Port: getEnv("AUTH_SERVICE_PORT", "3001"),
			},
			Billing: ServiceConfig{
				Host: getEnv("BILLING_SERVICE_HOST", "localhost"),
				Port: getEnv("BILLING_SERVICE_PORT", "3002"),
			},
			Analytics: ServiceConfig{
				Host: getEnv("ANALYTICS_SERVICE_HOST", "localhost"),
				Port: getEnv("ANALYTICS_SERVICE_PORT", "3003"),
			},
			Tenant: ServiceConfig{
				Host: getEnv("TENANT_SERVICE_HOST", "localhost"),
				Port: getEnv("TENANT_SERVICE_PORT", "3004"),
			},
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		},
		Stripe: StripeConfig{
			APIKey:        getEnv("STRIPE_API_KEY", ""),
			WebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		},
		RateLimit: RateLimitConfig{
			Requests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			Window:   getEnvAsDuration("RATE_LIMIT_WINDOW", 1*time.Minute),
		},
	}

	return cfg, nil
}

// DatabaseURL returns the PostgreSQL connection string
func (c *DatabaseConfig) DatabaseURL() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// RedisAddr returns the Redis address
func (c *RedisConfig) RedisAddr() string {
	if c.Password != "" {
		return fmt.Sprintf("%s:%s", c.Host, c.Port)
	}
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// Helper functions
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	var val int
	if _, err := fmt.Sscanf(valStr, "%d", &val); err != nil {
		return defaultVal
	}
	return val
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := time.ParseDuration(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func getEnvAsSlice(key string, defaultVal []string) []string {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	// Simple comma-separated parsing
	return splitAndTrim(valStr, ",")
}

func splitAndTrim(s string, sep string) []string {
	if s == "" {
		return []string{}
	}
	parts := make([]string, 0)
	for _, part := range split(s, sep) {
		if trimmed := trimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func split(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	var parts []string
	start := 0
	for i := 0; i < len(s); {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			parts = append(parts, s[start:i])
			start = i + len(sep)
			i += len(sep)
		} else {
			i++
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

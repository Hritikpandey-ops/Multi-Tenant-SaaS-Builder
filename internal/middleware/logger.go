package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Logger       *zerolog.Logger
	SkipPaths    []string
	UTC          bool
}

// DefaultLoggerConfig creates default logger configuration
func DefaultLoggerConfig(logger *zerolog.Logger) LoggerConfig {
	return LoggerConfig{
		Logger:    logger,
		SkipPaths: []string{"/health", "/metrics", "/ping"},
		UTC:       true,
	}
}

// Logger creates a structured logging middleware
func Logger(config LoggerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for certain paths
		for _, skipPath := range config.SkipPaths {
			if c.Request.URL.Path == skipPath {
				c.Next()
				return
			}
		}

		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Build log entry
		event := config.Logger.Info()

		// Add tenant context if available
		if tenantID, ok := GetTenantID(c); ok {
			event = event.Str("tenant_id", tenantID)
		}
		if userID, ok := GetUserID(c); ok {
			event = event.Str("user_id", userID)
		}

		// Calculate latency
		latency := time.Since(start)
		if config.UTC {
			event = event.Timestamp()
		}

		// Build log message
		status := c.Writer.Status()
		method := c.Request.Method
		ip := c.ClientIP()
		userAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		errors := c.Errors.String()

		if errors != "" {
			event = event.Str("errors", errors)
		}

		event.Str("client_ip", ip).
			Str("method", method).
			Str("path", path).
			Str("query", query).
			Int("status", status).
			Dur("latency", latency).
			Str("user_agent", userAgent).
			Str("referer", referer).
			Msg("HTTP Request")
	}
}

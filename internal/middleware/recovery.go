package middleware

import (
	"net/http"
	"runtime/debug"

	"Desktop/multitenant/saas/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// RecoveryConfig holds recovery middleware configuration
type RecoveryConfig struct {
	Logger     *zerolog.Logger
	StackTrace bool
}

// DefaultRecoveryConfig creates default recovery configuration
func DefaultRecoveryConfig(logger *zerolog.Logger) RecoveryConfig {
	return RecoveryConfig{
		Logger:     logger,
		StackTrace: true,
	}
}

// Recovery creates a panic recovery middleware
func Recovery(config RecoveryConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				event := config.Logger.Error().
					Interface("error", err).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Str("client_ip", c.ClientIP())

				// Add tenant context if available
				if tenantID, ok := GetTenantID(c); ok {
					event = event.Str("tenant_id", tenantID)
				}
				if userID, ok := GetUserID(c); ok {
					event = event.Str("user_id", userID)
				}

				// Add stack trace if enabled
				if config.StackTrace {
					event = event.Str("stack", string(debug.Stack()))
				}

				event.Msg("Panic recovered")

				// Send error response
				c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
					types.ErrCodeInternal,
					"Internal server error",
					nil,
				))
				c.Abort()
			}
		}()

		c.Next()
	}
}

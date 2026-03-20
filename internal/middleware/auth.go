package middleware

import (
	"net/http"
	"strings"

	"Desktop/multitenant/saas/internal/jwt"
	"Desktop/multitenant/saas/internal/types"

	"github.com/gin-gonic/gin"
)

const (
	// TenantIDKey is the context key for tenant ID
	TenantIDKey = "tenant_id"
	// UserIDKey is the context key for user ID
	UserIDKey = "user_id"
	// UserRoleKey is the context key for user role
	UserRoleKey = "user_role"
	// UserEmailKey is the context key for user email
	UserEmailKey = "user_email"
)

// AuthMiddleware creates a JWT authentication middleware
func AuthMiddleware(jwtManager *jwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
				types.ErrCodeUnauthorized,
				"Missing authorization header",
				nil,
			))
			c.Abort()
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
				types.ErrCodeUnauthorized,
				"Invalid authorization header format",
				nil,
			))
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
				types.ErrCodeUnauthorized,
				"Invalid or expired token",
				err.Error(),
			))
			c.Abort()
			return
		}

		// Set context values
		c.Set(TenantIDKey, claims.TenantID)
		c.Set(UserIDKey, claims.Sub)
		c.Set(UserRoleKey, claims.Role)
		c.Set(UserEmailKey, claims.Email)

		c.Next()
	}
}

// RequireRole checks if the user has the required role
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(UserRoleKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
				types.ErrCodeUnauthorized,
				"User not authenticated",
				nil,
			))
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		for _, allowedRole := range allowedRoles {
			if roleStr == allowedRole {
				c.Next()
				return
			}
		}

		// Check if user is OWNER (has all permissions)
		if roleStr == types.RoleOwner {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			types.ErrCodeForbidden,
			"Insufficient permissions",
			nil,
		))
		c.Abort()
	}
}

// GetTenantID retrieves the tenant ID from the context
func GetTenantID(c *gin.Context) (string, bool) {
	tenantID, exists := c.Get(TenantIDKey)
	if !exists {
		return "", false
	}
	return tenantID.(string), true
}

// GetUserID retrieves the user ID from the context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetUserRole retrieves the user role from the context
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(UserRoleKey)
	if !exists {
		return "", false
	}
	return role.(string), true
}

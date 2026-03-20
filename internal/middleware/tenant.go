package middleware

import (
	"context"
	"database/sql"
	"fmt"

	"Desktop/multitenant/saas/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// TenantDBMiddleware sets the tenant context for database queries
// This ensures that RLS (Row Level Security) policies work correctly
func TenantDBMiddleware(db *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := GetTenantID(c)
		if !exists {
			c.JSON(400, gin.H{"error": "Tenant ID not found in context"})
			c.Abort()
			return
		}

		// Set the tenant context for the database connection
		// This enables RLS policies that filter by tenant_id
		_, err := db.Exec(context.Background(), "SELECT set_tenant_context($1)", tenantID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to set tenant context"})
			c.Abort()
			return
		}

		// Store in context for potential use in handlers
		c.Set("db_tenant_set", true)

		c.Next()
	}
}

// GetTenantContext retrieves the full tenant context from Gin context
func GetTenantContext(c *gin.Context) (*types.TenantContext, error) {
	tenantID, ok := GetTenantID(c)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	userID, ok := GetUserID(c)
	if !ok {
		return nil, fmt.Errorf("user ID not found in context")
	}

	role, ok := GetUserRole(c)
	if !ok {
		return nil, fmt.Errorf("user role not found in context")
	}

	email, exists := c.Get(UserEmailKey)
	if !exists {
		return nil, fmt.Errorf("user email not found in context")
	}

	return &types.TenantContext{
		TenantID: tenantID,
		UserID:   userID,
		Role:     role,
		Email:    email.(string),
	}, nil
}

// WithTenantContext is a helper for services that need tenant context
func WithTenantContext(ctx context.Context, tenantID string) (context.Context, error) {
	// For sql.DB context
	return ctx, nil
}

// SetTenantContextDB sets the tenant context for a sql.DB transaction
func SetTenantContextDB(tx *sql.Tx, tenantID string) error {
	_, err := tx.Exec("SELECT set_tenant_context($1)", tenantID)
	return err
}

// SetTenantContextPGX sets the tenant context for a pgx connection
func SetTenantContextPGX(ctx context.Context, db PGXExecutor, tenantID string) error {
	_, err := db.Exec(ctx, "SELECT set_tenant_context($1)", tenantID)
	return err
}

// PGXExecutor is an interface for pgx execution
type PGXExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (interface{}, error)
}

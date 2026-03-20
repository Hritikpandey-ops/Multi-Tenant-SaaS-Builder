package middleware

import (
	"Desktop/multitenant/saas/internal/types"

	"github.com/gin-gonic/gin"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateJSON validates JSON request body
func ValidateJSON(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(obj); err != nil {
			c.JSON(400, types.NewErrorResponse(
				types.ErrCodeValidation,
				"Validation failed",
				gin.H{"errors": formatValidationErrors(err)},
			))
			c.Abort()
			return
		}
		c.Next()
	}
}

// ValidateQuery validates query parameters
func ValidateQuery(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindQuery(obj); err != nil {
			c.JSON(400, types.NewErrorResponse(
				types.ErrCodeValidation,
				"Validation failed",
				gin.H{"errors": formatValidationErrors(err)},
			))
			c.Abort()
			return
		}
		c.Next()
	}
}

// ValidateURI validates URI parameters
func ValidateURI(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindUri(obj); err != nil {
			c.JSON(400, types.NewErrorResponse(
				types.ErrCodeValidation,
				"Validation failed",
				gin.H{"errors": formatValidationErrors(err)},
			))
			c.Abort()
			return
		}
		c.Next()
	}
}

// formatValidationErrors formats validation errors
func formatValidationErrors(err error) []ValidationError {
	// This is a simplified version
	// In production, use a proper validation library like go-playground/validator
	// and format the errors properly
	return []ValidationError{
		{
			Field:   "request",
			Message: err.Error(),
		},
	}
}

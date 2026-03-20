package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"Desktop/multitenant/saas/internal/middleware"
	"Desktop/multitenant/saas/internal/repository"
	"Desktop/multitenant/saas/internal/types"
)

// UserHandler handles user management requests
type UserHandler struct {
	userRepo *repository.UserRepository
	logger   *zerolog.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userRepo *repository.UserRepository, logger *zerolog.Logger) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
		logger:   logger,
	}
}

// ListUsers returns a paginated list of users for the current tenant
func (h *UserHandler) ListUsers(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	ctx := c.Request.Context()

	users, total, err := h.userRepo.GetByTenant(ctx, tenantID, offset, limit)
	if err != nil {
		h.logger.Error().Err(err).Str("tenant_id", tenantID).Msg("Failed to list users")
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Failed to list users",
			nil,
		))
		return
	}

	// Convert to response format
	userResponses := make([]types.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = userToResponse(&user)
	}

	c.JSON(http.StatusOK, types.NewPaginatedResponse(userResponses, page, limit, int(total)))
}

// GetUser returns a specific user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)
	userID := c.Param("id")

	// Validate UUID
	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			types.ErrCodeValidation,
			"Invalid user ID",
			nil,
		))
		return
	}

	ctx := c.Request.Context()

	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, types.NewErrorResponse(
				types.ErrCodeNotFound,
				"User not found",
				nil,
			))
			return
		}

		h.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get user")
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Failed to get user",
			nil,
		))
		return
	}

	// Ensure user belongs to the tenant
	if user.TenantID != tenantID {
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			types.ErrCodeForbidden,
			"Access denied",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(userToResponse(user)))
}

// CreateUser creates a new user in the current tenant
func (h *UserHandler) CreateUser(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)
	currentUserID, _ := middleware.GetUserID(c)
	currentUserRole, _ := middleware.GetUserRole(c)

	// Check permissions: only OWNER and ADMIN can create users
	if currentUserRole != types.RoleOwner && currentUserRole != types.RoleAdmin {
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			types.ErrCodeForbidden,
			"Insufficient permissions to create users",
			nil,
		))
		return
	}

	var req types.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			types.ErrCodeValidation,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// For simplicity, this endpoint is for inviting users
	// In production, send invitation email
	h.logger.Info().
		Str("tenant_id", tenantID).
		Str("created_by", currentUserID).
		Msg("User creation requested")

	c.JSON(http.StatusNotImplemented, gin.H{"message": "User invitation not implemented yet"})
}

// UpdateUser updates an existing user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)
	currentUserID, _ := middleware.GetUserID(c)
	currentUserRole, _ := middleware.GetUserRole(c)
	userID := c.Param("id")

	// Validate UUID
	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			types.ErrCodeValidation,
			"Invalid user ID",
			nil,
		))
		return
	}

	ctx := c.Request.Context()

	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			types.ErrCodeNotFound,
			"User not found",
			nil,
		))
		return
	}

	// Check permissions: user can update their own profile, or OWNER/ADMIN can update anyone
	if userID != currentUserID && currentUserRole != types.RoleOwner && currentUserRole != types.RoleAdmin {
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			types.ErrCodeForbidden,
			"Insufficient permissions",
			nil,
		))
		return
	}

	// Ensure user belongs to the tenant
	if user.TenantID != tenantID {
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			types.ErrCodeForbidden,
			"Access denied",
			nil,
		))
		return
	}

	var req types.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			types.ErrCodeValidation,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Role != "" && (currentUserRole == types.RoleOwner || currentUserRole == types.RoleAdmin) {
		user.Role = req.Role
	}
	if req.Status != "" && (currentUserRole == types.RoleOwner || currentUserRole == types.RoleAdmin) {
		user.Status = req.Status
	}

	if err := h.userRepo.Update(ctx, user); err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update user")
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Failed to update user",
			nil,
		))
		return
	}

	h.logger.Info().
		Str("user_id", userID).
		Str("updated_by", currentUserID).
		Msg("User updated")

	c.JSON(http.StatusOK, types.NewSuccessResponse(userToResponse(user)))
}

// DeleteUser deletes (soft deletes) a user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)
	currentUserID, _ := middleware.GetUserID(c)
	currentUserRole, _ := middleware.GetUserRole(c)
	userID := c.Param("id")

	// Only OWNER and ADMIN can delete users
	if currentUserRole != types.RoleOwner && currentUserRole != types.RoleAdmin {
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			types.ErrCodeForbidden,
			"Insufficient permissions to delete users",
			nil,
		))
		return
	}

	// Cannot delete yourself
	if userID == currentUserID {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			types.ErrCodeValidation,
			"Cannot delete your own account",
			nil,
		))
		return
	}

	ctx := c.Request.Context()

	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			types.ErrCodeNotFound,
			"User not found",
			nil,
		))
		return
	}

	// Ensure user belongs to the tenant
	if user.TenantID != tenantID {
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			types.ErrCodeForbidden,
			"Access denied",
			nil,
		))
		return
	}

	if err := h.userRepo.Delete(ctx, tenantID, userID); err != nil {
		h.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to delete user")
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Failed to delete user",
			nil,
		))
		return
	}

	h.logger.Info().
		Str("user_id", userID).
		Str("deleted_by", currentUserID).
		Msg("User deleted")

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// InviteUser invites a new user to the tenant
func (h *UserHandler) InviteUser(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)
	currentUserID, _ := middleware.GetUserID(c)

	var req types.InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			types.ErrCodeValidation,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	h.logger.Info().
		Str("tenant_id", tenantID).
		Str("invited_by", currentUserID).
		Str("email", req.Email).
		Str("role", req.Role).
		Msg("User invitation requested")

	// In production, send invitation email with magic link
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "User invitation will be sent via email",
		"email":   req.Email,
	})
}

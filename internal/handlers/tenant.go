package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"Desktop/multitenant/saas/internal/middleware"
	"Desktop/multitenant/saas/internal/repository"
	"Desktop/multitenant/saas/internal/types"
)

// TenantHandler handles tenant management requests
type TenantHandler struct {
	tenantRepo *repository.TenantRepository
	logger     *zerolog.Logger
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantRepo *repository.TenantRepository, logger *zerolog.Logger) *TenantHandler {
	return &TenantHandler{
		tenantRepo: tenantRepo,
		logger:     logger,
	}
}

// GetTenant returns the current tenant's information
func (h *TenantHandler) GetTenant(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)

	ctx := c.Request.Context()

	tenant, err := h.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			types.ErrCodeNotFound,
			"Tenant not found",
			nil,
		))
		return
	}

	// Get user count
	userCount, _ := h.tenantRepo.GetUserCount(ctx, tenantID)

	response := gin.H{
		"id":         tenant.ID,
		"name":       tenant.Name,
		"slug":       tenant.Slug,
		"plan":       tenant.Plan,
		"status":     tenant.Status,
		"user_count": userCount,
		"created_at": tenant.CreatedAt,
		"updated_at": tenant.UpdatedAt,
	}

	if tenant.Metadata != nil {
		response["metadata"] = tenant.Metadata
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(response))
}

// UpdateTenant updates the current tenant's information
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)
	currentUserRole, _ := middleware.GetUserRole(c)

	// Only OWNER can update tenant settings
	if currentUserRole != types.RoleOwner {
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			types.ErrCodeForbidden,
			"Only owners can update tenant settings",
			nil,
		))
		return
	}

	var req types.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			types.ErrCodeValidation,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()

	tenant, err := h.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			types.ErrCodeNotFound,
			"Tenant not found",
			nil,
		))
		return
	}

	// Update fields
	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.Status != "" {
		tenant.Status = req.Status
	}

	if err := h.tenantRepo.Update(ctx, tenant); err != nil {
		h.logger.Error().Err(err).Str("tenant_id", tenantID).Msg("Failed to update tenant")
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Failed to update tenant",
			nil,
		))
		return
	}

	h.logger.Info().
		Str("tenant_id", tenantID).
		Msg("Tenant updated")

	c.JSON(http.StatusOK, types.NewSuccessResponse(gin.H{
		"id":         tenant.ID,
		"name":       tenant.Name,
		"slug":       tenant.Slug,
		"plan":       tenant.Plan,
		"status":     tenant.Status,
		"created_at": tenant.CreatedAt,
		"updated_at": tenant.UpdatedAt,
	}))
}

// GetTenantUsers returns all users for the current tenant
func (h *TenantHandler) GetTenantUsers(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)

	ctx := c.Request.Context()

	users, err := h.tenantRepo.GetUsers(ctx, tenantID)
	if err != nil {
		h.logger.Error().Err(err).Str("tenant_id", tenantID).Msg("Failed to get tenant users")
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Failed to get users",
			nil,
		))
		return
	}

	userResponses := make([]types.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = userToResponse(&user)
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(gin.H{
		"users": userResponses,
		"count": len(userResponses),
	}))
}

// GetTenantUsage returns usage statistics for the tenant
func (h *TenantHandler) GetTenantUsage(c *gin.Context) {
	tenantID, _ := middleware.GetTenantID(c)

	ctx := c.Request.Context()

	tenant, err := h.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			types.ErrCodeNotFound,
			"Tenant not found",
			nil,
		))
		return
	}

	userCount, _ := h.tenantRepo.GetUserCount(ctx, tenantID)

	// Get plan limits
	planLimits := getPlanLimits(tenant.Plan)

	usage := gin.H{
		"tenant_id":        tenantID,
		"plan":             tenant.Plan,
		"user_count":       userCount,
		"user_limit":       planLimits.MaxUsers,
		"usage_percentage": gin.H{},
	}

	if planLimits.MaxUsers > 0 {
		usage["usage_percentage"].(gin.H)["users"] = float64(userCount) / float64(planLimits.MaxUsers) * 100
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(usage))
}

// PlanLimits represents the limits for each plan
type PlanLimits struct {
	MaxUsers  int
	StorageMB int
}

func getPlanLimits(plan string) PlanLimits {
	switch plan {
	case types.PlanFree:
		return PlanLimits{MaxUsers: 5, StorageMB: 100}
	case types.PlanPro:
		return PlanLimits{MaxUsers: 50, StorageMB: 10000}
	case types.PlanEnterprise:
		return PlanLimits{MaxUsers: -1, StorageMB: -1} // Unlimited
	default:
		return PlanLimits{MaxUsers: 5, StorageMB: 100}
	}
}

package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	"Desktop/multitenant/saas/internal/database"
	"Desktop/multitenant/saas/internal/jwt"
	"Desktop/multitenant/saas/internal/middleware"
	"Desktop/multitenant/saas/internal/repository"
	"Desktop/multitenant/saas/internal/types"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	userRepo   *repository.UserRepository
	tenantRepo *repository.TenantRepository
	jwtManager *jwt.JWTManager
	redis      *database.RedisClient
	logger     *zerolog.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userRepo *repository.UserRepository,
	tenantRepo *repository.TenantRepository,
	jwtManager *jwt.JWTManager,
	redis *database.RedisClient,
	logger *zerolog.Logger,
) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		jwtManager: jwtManager,
		redis:      redis,
		logger:     logger,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req types.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			types.ErrCodeValidation,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()

	// Check if this is the first user (creating new tenant)
	// or an existing user joining a tenant
	var tenant *types.Tenant
	var err error

	if req.TenantName != "" && req.TenantSlug != "" {
		// Create new tenant
		tenant = &types.Tenant{
			Name:   req.TenantName,
			Slug:   req.TenantSlug,
			Plan:   types.PlanFree,
			Status: types.TenantStatusActive,
		}

		if err := h.tenantRepo.Create(ctx, tenant); err != nil {
			h.logger.Error().Err(err).Str("slug", req.TenantSlug).Msg("Failed to create tenant")
			c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
				types.ErrCodeInternal,
				"Failed to create tenant",
				nil,
			))
			return
		}

		h.logger.Info().Str("tenant_id", tenant.ID).Str("name", tenant.Name).Msg("New tenant created")
	} else {
		// For now, require tenant info
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			types.ErrCodeValidation,
			"Tenant information required",
			"tenant_name and tenant_slug are required",
		))
		return
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to hash password")
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Failed to process password",
			nil,
		))
		return
	}

	// Create user
	user := &types.User{
		TenantID:      tenant.ID,
		Email:         req.Email,
		PasswordHash:  string(passwordHash),
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Role:          types.RoleOwner, // First user is owner
		Status:        types.UserStatusActive,
		EmailVerified: false, // In production, send verification email
	}

	if err := h.userRepo.Create(ctx, user); err != nil {
		h.logger.Error().Err(err).Str("email", req.Email).Msg("Failed to create user")
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Failed to create user",
			nil,
		))
		return
	}

	// Generate JWT token
	tokenDetails, err := h.jwtManager.GenerateToken(user, tenant)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate token")
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Failed to generate token",
			nil,
		))
		return
	}

	// Store refresh token in Redis (with expiration)
	if err := h.redis.Set(
		ctx,
		"refresh_token:"+tokenDetails.RefreshToken,
		user.ID,
		time.Hour*24*7, // 7 days
	).Err(); err != nil {
		h.logger.Error().Err(err).Msg("Failed to store refresh token")
	}

	// Update last login
	_ = h.userRepo.UpdateLastLogin(ctx, user.ID)

	h.logger.Info().
		Str("user_id", user.ID).
		Str("tenant_id", tenant.ID).
		Str("email", user.Email).
		Msg("User registered successfully")

	c.JSON(http.StatusCreated, types.AuthResponse{
		Token:        tokenDetails.Token,
		RefreshToken: tokenDetails.RefreshToken,
		ExpiresIn:    tokenDetails.ExpiresIn,
		User:         userToResponse(user),
		Tenant:       tenantToResponse(tenant),
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req types.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			types.ErrCodeValidation,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()

	// Find user by email
	// For multi-tenant login, we need tenant identifier
	// For now, we'll search by email across all tenants
	// In production, require tenant slug or subdomain
	user, err := h.findUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
				types.ErrCodeUnauthorized,
				"Invalid credentials",
				nil,
			))
			return
		}

		h.logger.Error().Err(err).Str("email", req.Email).Msg("Failed to find user")
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Login failed",
			nil,
		))
		return
	}

	// Check if user is active
	if user.Status != types.UserStatusActive {
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			types.ErrCodeForbidden,
			"User account is not active",
			nil,
		))
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			types.ErrCodeUnauthorized,
			"Invalid credentials",
			nil,
		))
		return
	}

	// Get tenant
	tenant, err := h.tenantRepo.GetByID(ctx, user.TenantID)
	if err != nil {
		h.logger.Error().Err(err).Str("tenant_id", user.TenantID).Msg("Failed to get tenant")
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Login failed",
			nil,
		))
		return
	}

	// Check if tenant is active
	if tenant.Status != types.TenantStatusActive {
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			types.ErrCodeTenant,
			"Tenant account is not active",
			nil,
		))
		return
	}

	// Generate JWT token
	tokenDetails, err := h.jwtManager.GenerateToken(user, tenant)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate token")
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Failed to generate token",
			nil,
		))
		return
	}

	// Store refresh token in Redis
	if err := h.redis.Set(
		ctx,
		"refresh_token:"+tokenDetails.RefreshToken,
		user.ID,
		time.Hour*24*7,
	).Err(); err != nil {
		h.logger.Error().Err(err).Msg("Failed to store refresh token")
	}

	// Update last login
	_ = h.userRepo.UpdateLastLogin(ctx, user.ID)

	h.logger.Info().
		Str("user_id", user.ID).
		Str("tenant_id", tenant.ID).
		Str("email", user.Email).
		Msg("User logged in successfully")

	c.JSON(http.StatusOK, types.AuthResponse{
		Token:        tokenDetails.Token,
		RefreshToken: tokenDetails.RefreshToken,
		ExpiresIn:    tokenDetails.ExpiresIn,
		User:         userToResponse(user),
		Tenant:       tenantToResponse(tenant),
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req types.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			types.ErrCodeValidation,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()

	// Validate refresh token and get user ID
	userID, err := h.validateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			types.ErrCodeUnauthorized,
			"Invalid refresh token",
			nil,
		))
		return
	}

	// Get user and tenant
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			types.ErrCodeUnauthorized,
			"User not found",
			nil,
		))
		return
	}

	tenant, err := h.tenantRepo.GetByID(ctx, user.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Failed to get tenant",
			nil,
		))
		return
	}

	// Generate new access token
	tokenDetails, err := h.jwtManager.GenerateToken(user, tenant)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Failed to generate token",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, types.AuthResponse{
		Token:        tokenDetails.Token,
		RefreshToken: req.RefreshToken, // Return same refresh token
		ExpiresIn:    tokenDetails.ExpiresIn,
		User:         userToResponse(user),
		Tenant:       tenantToResponse(tenant),
	})
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	tenantID, _ := middleware.GetTenantID(c)

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

	tenant, err := h.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Failed to get tenant",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":   userToResponse(user),
		"tenant": tenantToResponse(tenant),
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a production system, add the token to a blacklist in Redis
	// For now, we just return success (client should discard the token)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			types.ErrCodeValidation,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()

	// Get user
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			types.ErrCodeNotFound,
			"User not found",
			nil,
		))
		return
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			types.ErrCodeUnauthorized,
			"Invalid current password",
			nil,
		))
		return
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Failed to process password",
			nil,
		))
		return
	}

	// Update password
	if err := h.userRepo.UpdatePassword(ctx, userID, string(newPasswordHash)); err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			types.ErrCodeInternal,
			"Failed to update password",
			nil,
		))
		return
	}

	h.logger.Info().Str("user_id", userID).Msg("Password changed successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// Helper functions

func (h *AuthHandler) findUserByEmail(ctx context.Context, email string) (*types.User, error) {
	// Try to find user by email - this searches across all tenants
	// In production, you should require tenant slug/subdomain in the request
	return h.userRepo.FindByEmailAnyTenant(ctx, email)
}

func (h *AuthHandler) validateRefreshToken(ctx context.Context, token string) (string, error) {
	userID, err := h.redis.Get(ctx, "refresh_token:"+token).Result()
	if err != nil {
		return "", errors.New("invalid refresh token")
	}
	return userID, nil
}

func userToResponse(user *types.User) types.UserResponse {
	return types.UserResponse{
		ID:            user.ID,
		TenantID:      user.TenantID,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Role:          user.Role,
		Status:        user.Status,
		EmailVerified: user.EmailVerified,
		LastLoginAt:   user.LastLoginAt,
		CreatedAt:     user.CreatedAt,
	}
}

func tenantToResponse(tenant *types.Tenant) types.TenantResponse {
	return types.TenantResponse{
		ID:        tenant.ID,
		Name:      tenant.Name,
		Slug:      tenant.Slug,
		Plan:      tenant.Plan,
		Status:    tenant.Status,
		CreatedAt: tenant.CreatedAt,
		UpdatedAt: tenant.UpdatedAt,
	}
}

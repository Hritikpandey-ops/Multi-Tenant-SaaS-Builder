package jwt

import (
	"errors"
	"fmt"
	"time"

	"Desktop/multitenant/saas/internal/config"
	"Desktop/multitenant/saas/internal/types"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTManager handles JWT token creation and validation
type JWTManager struct {
	secretKey     []byte
	rsaPrivateKey *jwt.SigningMethodRSA
	rsaPublicKey  interface{}
	config        config.JWTConfig
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(cfg config.JWTConfig) (*JWTManager, error) {
	manager := &JWTManager{
		secretKey: []byte(cfg.Secret),
		config:    cfg,
	}

	// For production, load RSA keys if available
	// This enables RS256 algorithm which is more secure than HS256
	// For now, we'll use HS256 with the secret key

	return manager, nil
}

// GenerateToken generates a new JWT token for a user
func (j *JWTManager) GenerateToken(user *types.User, tenant *types.Tenant) (*types.TokenDetails, error) {
	now := time.Now()
	expiresAt := now.Add(j.config.ExpiresIn)

	claims := jwt.MapClaims{
		"sub":       user.ID,
		"tenant_id": tenant.ID,
		"email":     user.Email,
		"role":      user.Role,
		"iat":       now.Unix(),
		"exp":       expiresAt.Unix(),
		"iss":       j.config.Issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	// Generate refresh token
	refreshToken := uuid.New().String()

	return &types.TokenDetails{
		Token:        tokenString,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(j.config.ExpiresIn.Seconds()),
		ExpiresAt:    expiresAt,
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*types.JWTPayload, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, errors.New("token expired")
		}
	}

	// Extract claims
	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("invalid subject claim")
	}

	tenantID, ok := claims["tenant_id"].(string)
	if !ok {
		return nil, errors.New("invalid tenant_id claim")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New("invalid email claim")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("invalid role claim")
	}

	iat, _ := claims["iat"].(float64)
	exp, _ := claims["exp"].(float64)

	return &types.JWTPayload{
		Sub:       sub,
		TenantID:  tenantID,
		Email:     email,
		Role:      role,
		IssuedAt:  int64(iat),
		ExpiresAt: int64(exp),
	}, nil
}

// GenerateRefreshToken generates a new refresh token
func (j *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	return uuid.New().String(), nil
}

// ValidateRefreshToken validates a refresh token
// In production, this should check against a database or Redis
func (j *JWTManager) ValidateRefreshToken(refreshToken string) (string, error) {
	// For now, just validate it's a valid UUID
	_, err := uuid.Parse(refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	// In production, you would:
	// 1. Check if the refresh token exists in Redis/database
	// 2. Check if it's expired
	// 3. Return the associated user ID

	return "", nil
}

// RefreshAccessToken generates a new access token from a refresh token
func (j *JWTManager) RefreshAccessToken(refreshToken string, user *types.User, tenant *types.Tenant) (*types.TokenDetails, error) {
	// Validate refresh token
	userID, err := j.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// In production, verify that the user ID from the refresh token matches the provided user
	if userID != "" && userID != user.ID {
		return nil, errors.New("refresh token does not belong to user")
	}

	// Generate new access token
	return j.GenerateToken(user, tenant)
}

// ExtractTenantFromToken extracts tenant ID from token without full validation
// Useful for logging and metrics
func (j *JWTManager) ExtractTenantFromToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}

	tenantID, ok := claims["tenant_id"].(string)
	if !ok {
		return "", errors.New("no tenant_id in token")
	}

	return tenantID, nil
}

// WithTenantContext adds tenant context to JWT claims
func (j *JWTManager) WithTenantContext(baseClaims jwt.MapClaims, tenantID string) jwt.MapClaims {
	if baseClaims == nil {
		baseClaims = make(jwt.MapClaims)
	}
	baseClaims["tenant_id"] = tenantID
	return baseClaims
}

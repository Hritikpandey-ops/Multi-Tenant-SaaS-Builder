package types

import "time"

// JWTPayload represents the JWT token payload
type JWTPayload struct {
	Sub      string `json:"sub"`       // User ID
	TenantID string `json:"tenant_id"` // Tenant ID
	Email    string `json:"email"`
	Role     string `json:"role"`
	IssuedAt int64  `json:"iat"`
	ExpiresAt int64 `json:"exp"`
}

// TokenDetails represents token details
type TokenDetails struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TenantContext represents tenant context extracted from JWT
type TenantContext struct {
	TenantID string
	UserID   string
	Role     string
	Email    string
}

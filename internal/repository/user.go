package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"Desktop/multitenant/saas/internal/types"
)

// UserRepository handles user data operations
type UserRepository struct {
	DB *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *types.User) error {
	user.ID = uuid.New().String()
	return r.DB.WithContext(ctx).Create(user).Error
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*types.User, error) {
	var user types.User
	err := r.DB.WithContext(ctx).
		Preload("Tenant").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByEmail retrieves a user by email (tenant-specific)
func (r *UserRepository) GetByEmail(ctx context.Context, tenantID, email string) (*types.User, error) {
	var user types.User
	err := r.DB.WithContext(ctx).
		Preload("Tenant").
		Where("tenant_id = ? AND email = ? AND deleted_at IS NULL", tenantID, email).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByTenant retrieves all users for a tenant
func (r *UserRepository) GetByTenant(ctx context.Context, tenantID string, offset, limit int) ([]types.User, int64, error) {
	var users []types.User
	var total int64

	query := r.DB.WithContext(ctx).
		Model(&types.User{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *types.User) error {
	return r.DB.WithContext(ctx).
		Model(&types.User{}).
		Where("id = ? AND tenant_id = ?", user.ID, user.TenantID).
		Updates(user).Error
}

// Delete soft deletes a user
func (r *UserRepository) Delete(ctx context.Context, tenantID, id string) error {
	return r.DB.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&types.User{}).Error
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	return r.DB.WithContext(ctx).
		Model(&types.User{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash).Error
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	return r.DB.WithContext(ctx).
		Model(&types.User{}).
		Where("id = ?", userID).
		Update("last_login_at", gorm.Expr("NOW()")).Error
}

// SetEmailVerified sets the email verified flag
func (r *UserRepository) SetEmailVerified(ctx context.Context, userID string) error {
	return r.DB.WithContext(ctx).
		Model(&types.User{}).
		Where("id = ?", userID).
		Update("email_verified", true).Error
}

// FindByEmailAnyTenant finds a user by email across all tenants
// WARNING: This should only be used for login where tenant is identified separately
func (r *UserRepository) FindByEmailAnyTenant(ctx context.Context, email string) (*types.User, error) {
	var user types.User
	err := r.DB.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// Repository errors
var (
	ErrUserNotFound   = errors.New("user not found")
	ErrTenantNotFound = errors.New("tenant not found")
	ErrEmailExists    = errors.New("email already exists")
	ErrInvalidTenant  = errors.New("invalid tenant")
)

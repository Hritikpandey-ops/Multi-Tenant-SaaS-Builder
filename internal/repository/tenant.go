package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"Desktop/multitenant/saas/internal/types"
)

// TenantRepository handles tenant data operations
type TenantRepository struct {
	db *gorm.DB
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// Create creates a new tenant
func (r *TenantRepository) Create(ctx context.Context, tenant *types.Tenant) error {
	tenant.ID = uuid.New().String()
	return r.db.WithContext(ctx).Create(tenant).Error
}

// GetByID retrieves a tenant by ID
func (r *TenantRepository) GetByID(ctx context.Context, id string) (*types.Tenant, error) {
	var tenant types.Tenant
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&tenant).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}

	return &tenant, nil
}

// GetBySlug retrieves a tenant by slug
func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*types.Tenant, error) {
	var tenant types.Tenant
	err := r.db.WithContext(ctx).
		Where("slug = ? AND deleted_at IS NULL", slug).
		First(&tenant).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}

	return &tenant, nil
}

// List retrieves all tenants (admin only)
func (r *TenantRepository) List(ctx context.Context, offset, limit int) ([]types.Tenant, int64, error) {
	var tenants []types.Tenant
	var total int64

	query := r.db.WithContext(ctx).
		Model(&types.Tenant{}).
		Where("deleted_at IS NULL")

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&tenants).Error

	if err != nil {
		return nil, 0, err
	}

	return tenants, total, nil
}

// Update updates a tenant
func (r *TenantRepository) Update(ctx context.Context, tenant *types.Tenant) error {
	return r.db.WithContext(ctx).
		Model(&types.Tenant{}).
		Where("id = ?", tenant.ID).
		Updates(tenant).Error
}

// Delete soft deletes a tenant
func (r *TenantRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&types.Tenant{}).Error
}

// UpdateStatus updates a tenant's status
func (r *TenantRepository) UpdateStatus(ctx context.Context, id, status string) error {
	return r.db.WithContext(ctx).
		Model(&types.Tenant{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// GetUsers retrieves all users for a tenant
func (r *TenantRepository) GetUsers(ctx context.Context, tenantID string) ([]types.User, error) {
	var users []types.User
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Find(&users).Error

	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetUserCount returns the number of users in a tenant
func (r *TenantRepository) GetUserCount(ctx context.Context, tenantID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&types.User{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&count).Error

	return count, err
}

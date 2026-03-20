package types

import "time"

// Tenant represents a multi-tenant organization
type Tenant struct {
	ID        string    `json:"id" db:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name      string    `json:"name" db:"name" gorm:"not null" validate:"required,min=2,max=255"`
	Slug      string    `json:"slug" db:"slug" gorm:"uniqueIndex;not null" validate:"required,alphanum,min=2,max=100"`
	Plan      string    `json:"plan" db:"plan" gorm:"default:free" validate:"oneof=free pro enterprise"`
	Status    string    `json:"status" db:"status" gorm:"default:active" validate:"oneof=active suspended cancelled"`
	Metadata  JSONB     `json:"metadata,omitempty" db:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at" gorm:"index"`
}

// TenantStatus enum
const (
	TenantStatusActive     = "active"
	TenantStatusSuspended  = "suspended"
	TenantStatusCancelled  = "cancelled"
)

// TenantPlan enum
const (
	PlanFree      = "free"
	PlanPro       = "pro"
	PlanEnterprise = "enterprise"
)

// CreateTenantRequest represents a request to create a tenant
type CreateTenantRequest struct {
	Name string `json:"name" validate:"required,min=2,max=255"`
	Slug string `json:"slug" validate:"required,alphanum,min=2,max=100"`
	Plan string `json:"plan" validate:"omitempty,oneof=free pro enterprise"`
}

// UpdateTenantRequest represents a request to update a tenant
type UpdateTenantRequest struct {
	Name   string `json:"name" validate:"omitempty,min=2,max=255"`
	Status string `json:"status" validate:"omitempty,oneof=active suspended cancelled"`
}

// TenantResponse represents a tenant response
type TenantResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Plan      string    `json:"plan"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Tenant) TableName() string {
	return "tenants"
}

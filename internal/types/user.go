package types

import "time"

// User represents a user in the system
type User struct {
	ID           string     `json:"id" db:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	TenantID     string     `json:"tenant_id" db:"tenant_id" gorm:"not null;index:idx_users_tenant_id;index:idx_users_tenant_email" validate:"required,uuid"`
	Email        string     `json:"email" db:"email" gorm:"not null;index:idx_users_email;index:idx_users_tenant_email" validate:"required,email"`
	PasswordHash string     `json:"-" db:"password_hash" gorm:"not null"`
	FirstName    string     `json:"first_name" db:"first_name" validate:"omitempty,max=100"`
	LastName     string     `json:"last_name" db:"last_name" validate:"omitempty,max=100"`
	Role         string     `json:"role" db:"role" gorm:"default:MEMBER" validate:"oneof=OWNER ADMIN MEMBER VIEWER"`
	Status       string     `json:"status" db:"status" gorm:"default:active" validate:"oneof=active invited suspended"`
	EmailVerified bool     `json:"email_verified" db:"email_verified" gorm:"default:false"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at" gorm:"index"`

	// Relations
	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// UserRole enum
const (
	RoleOwner  = "OWNER"
	RoleAdmin  = "ADMIN"
	RoleMember = "MEMBER"
	RoleViewer = "VIEWER"
)

// UserStatus enum
const (
	UserStatusActive    = "active"
	UserStatusInvited   = "invited"
	UserStatusSuspended = "suspended"
)

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
	// For creating first user, optionally include tenant info
	TenantName string `json:"tenant_name,omitempty" validate:"omitempty,min=2,max=255"`
	TenantSlug string `json:"tenant_slug,omitempty" validate:"omitempty,alphanum,min=2,max=100"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	User         UserResponse `json:"user"`
	Tenant       TenantResponse `json:"tenant"`
}

// UserResponse represents a user response (without sensitive data)
type UserResponse struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	Email         string     `json:"email"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Role          string     `json:"role"`
	Status        string     `json:"status"`
	EmailVerified bool       `json:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	FirstName string `json:"first_name" validate:"omitempty,max=100"`
	LastName  string `json:"last_name" validate:"omitempty,max=100"`
	Role      string `json:"role" validate:"omitempty,oneof=OWNER ADMIN MEMBER VIEWER"`
	Status    string `json:"status" validate:"omitempty,oneof=active invited suspended"`
}

// InviteUserRequest represents a request to invite a user
type InviteUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Role     string `json:"role" validate:"required,oneof=OWNER ADMIN MEMBER VIEWER"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

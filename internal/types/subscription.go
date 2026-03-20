package types

import "time"

// Plan represents a subscription plan
type Plan struct {
	ID          string    `json:"id" db:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name        string    `json:"name" db:"name" gorm:"uniqueIndex;not null" validate:"required,max=100"`
	Slug        string    `json:"slug" db:"slug" gorm:"uniqueIndex;not null" validate:"required,alphanum,max=50"`
	Description string    `json:"description" db:"description"`
	PriceCents  int       `json:"price_cents" db:"price_cents" gorm:"default:0" validate:"min=0"`
	Currency    string    `json:"currency" db:"currency" gorm:"default:USD" validate:"required,len=3"`
	Interval    string    `json:"interval" db:"interval" gorm:"not null" validate:"required,oneof=monthly yearly"`
	Features    JSONB     `json:"features" db:"features" gorm:"type:jsonb;default:'[]'"`
	Limits      JSONB     `json:"limits" db:"limits" gorm:"type:jsonb;default:'{}'"`
	StripePriceID string  `json:"stripe_price_id,omitempty" db:"stripe_price_id"`
	IsActive    bool      `json:"is_active" db:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Subscription represents a tenant's subscription
type Subscription struct {
	ID                 string     `json:"id" db:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	TenantID           string     `json:"tenant_id" db:"tenant_id" gorm:"not null;uniqueIndex;index:idx_subscriptions_tenant_id" validate:"required,uuid"`
	PlanID             string     `json:"plan_id" db:"plan_id" gorm:"not null" validate:"required,uuid"`
	Status             string     `json:"status" db:"status" gorm:"default:incomplete" validate:"required,oneof=active past_due cancelled incomplete incomplete_expired trialing unpaid"`
	StripeCustomerID   string     `json:"stripe_customer_id,omitempty" db:"stripe_customer_id"`
	StripeSubscriptionID string   `json:"stripe_subscription_id,omitempty" db:"stripe_subscription_id"`
	CurrentPeriodStart *time.Time `json:"current_period_start,omitempty" db:"current_period_start"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty" db:"current_period_end"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end" db:"cancel_at_period_end" gorm:"default:false"`
	CanceledAt         *time.Time `json:"canceled_at,omitempty" db:"canceled_at"`
	TrialStart         *time.Time `json:"trial_start,omitempty" db:"trial_start"`
	TrialEnd           *time.Time `json:"trial_end,omitempty" db:"trial_end"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`

	// Relations
	Plan  *Plan   `json:"plan,omitempty" gorm:"foreignKey:PlanID"`
	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// Subscription status constants
const (
	SubscriptionStatusActive           = "active"
	SubscriptionStatusPastDue         = "past_due"
	SubscriptionStatusCancelled       = "cancelled"
	SubscriptionStatusIncomplete      = "incomplete"
	SubscriptionStatusIncompleteExpired = "incomplete_expired"
	SubscriptionStatusTrialing        = "trialing"
	SubscriptionStatusUnpaid          = "unpaid"
)

// Plan interval constants
const (
	IntervalMonthly = "monthly"
	IntervalYearly  = "yearly"
)

// CreateSubscriptionRequest represents a request to create a subscription
type CreateSubscriptionRequest struct {
	PlanID string `json:"plan_id" validate:"required,uuid"`
}

// UpdateSubscriptionRequest represents a request to update a subscription
type UpdateSubscriptionRequest struct {
	PlanID            string `json:"plan_id" validate:"omitempty,uuid"`
	CancelAtPeriodEnd bool   `json:"cancel_at_period_end"`
}

// SubscriptionResponse represents a subscription response
type SubscriptionResponse struct {
	ID                 string     `json:"id"`
	TenantID           string     `json:"tenant_id"`
	Plan               PlanResponse `json:"plan"`
	Status             string     `json:"status"`
	CurrentPeriodStart *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end"`
	CanceledAt         *time.Time `json:"canceled_at,omitempty"`
	TrialStart         *time.Time `json:"trial_start,omitempty"`
	TrialEnd           *time.Time `json:"trial_end,omitempty"`
}

// PlanResponse represents a plan response
type PlanResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	PriceCents  int      `json:"price_cents"`
	Currency    string   `json:"currency"`
	Interval    string   `json:"interval"`
	Features    []string `json:"features"`
	Limits      map[string]interface{} `json:"limits"`
	IsActive    bool     `json:"is_active"`
}

// TableName specifies the table name for GORM
func (Plan) TableName() string {
	return "plans"
}

func (Subscription) TableName() string {
	return "subscriptions"
}

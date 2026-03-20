package types

import "time"

// UsageEvent represents a usage/event record
type UsageEvent struct {
	ID        string    `json:"id" db:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	TenantID  string    `json:"tenant_id" db:"tenant_id" gorm:"not null;index:idx_usage_events_tenant_id" validate:"required,uuid"`
	UserID    *string   `json:"user_id,omitempty" db:"user_id" gorm:"index;index:idx_usage_events_timestamp" validate:"omitempty,uuid"`
	EventType string    `json:"event_type" db:"event_type" gorm:"not null;index:idx_usage_events_type" validate:"required,max=100"`
	EventName string    `json:"event_name" db:"event_name" gorm:"not null" validate:"required,max=255"`
	Properties JSONB    `json:"properties,omitempty" db:"properties" gorm:"type:jsonb;default:'{}'"`
	Timestamp time.Time `json:"timestamp" db:"timestamp" gorm:"not null;index:idx_usage_events_timestamp"`
	Metadata  JSONB     `json:"metadata,omitempty" db:"metadata" gorm:"type:jsonb;default:'{}'"`

	// Relations
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// RecordEventRequest represents a request to record an event
type RecordEventRequest struct {
	EventType string                 `json:"event_type" validate:"required,max=100"`
	EventName string                 `json:"event_name" validate:"required,max=255"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp *time.Time             `json:"timestamp,omitempty"`
}

// RecordEventsRequest represents a batch request to record multiple events
type RecordEventsRequest struct {
	Events []RecordEventRequest `json:"events" validate:"required,min=1,max=100"`
}

// AnalyticsQueryRequest represents a request to query analytics
type AnalyticsQueryRequest struct {
	EventType string    `json:"event_type,omitempty" validate:"omitempty,max=100"`
	EventName string    `json:"event_name,omitempty" validate:"omitempty,max=255"`
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required,gt=StartDate"`
	GroupBy   string    `json:"group_by,omitempty" validate:"omitempty,oneof=day week month hour"`
}

// AnalyticsResponse represents aggregated analytics
type AnalyticsResponse struct {
	TotalEvents   int64                `json:"total_events"`
	EventType     string               `json:"event_type,omitempty"`
	StartDate     time.Time            `json:"start_date"`
	EndDate       time.Time            `json:"end_date"`
	TimeSeries    []TimeSeriesDataPoint `json:"time_series,omitempty"`
	TopEvents     []EventCount         `json:"top_events,omitempty"`
}

// TimeSeriesDataPoint represents a data point in a time series
type TimeSeriesDataPoint struct {
	Timestamp string `json:"timestamp"`
	Count     int64  `json:"count"`
}

// EventCount represents an event count
type EventCount struct {
	EventName string `json:"event_name"`
	Count     int64  `json:"count"`
}

// TableName specifies the table name for GORM
func (UsageEvent) TableName() string {
	return "usage_events"
}

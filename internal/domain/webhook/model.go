package webhook

import (
	"time"

	"github.com/lib/pq"
	"github.com/turtacn/QuantaID/pkg/types"
)

// Subscription represents a webhook subscription.
// We use pq.StringArray to support Postgres array types.
// Although this introduces a slight coupling to Postgres driver types in the domain,
// it is the most efficient way to handle array columns with GORM/Postgres.
// An alternative would be a custom Scanner/Valuer or a separate join table, but for tags/events lists,
// arrays are idiomatic in Postgres.
type Subscription struct {
	ID        string         `json:"id" gorm:"primaryKey"`
	URL       string         `json:"url"`
	Secret    string         `json:"-"` // Encrypted storage
	Events    pq.StringArray `json:"events" gorm:"type:text[]"`
	Active    bool           `json:"active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type DeliveryTask struct {
	SubscriptionID string
	EventID        string
	EventType      string
	Payload        interface{}
	Attempt        int
	NextRetry      time.Time
}

// Repository defines the interface for webhook subscription storage.
type Repository interface {
	Create(subscription *Subscription) error
	GetByID(id string) (*Subscription, error)
	List() ([]*Subscription, error)
	Update(subscription *Subscription) error
	Delete(id string) error
	FindByEventType(eventType string) ([]*Subscription, error)
}

// Ensure the Subscription is compatible with GORM
func (Subscription) TableName() string {
	return "webhook_subscriptions"
}

// ErrSubscriptionNotFound is returned when a subscription is not found.
var ErrSubscriptionNotFound = &types.Error{
	Code:    "SUBSCRIPTION_NOT_FOUND",
	Message: "webhook subscription not found",
}

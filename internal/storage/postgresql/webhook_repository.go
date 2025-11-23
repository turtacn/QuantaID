package postgresql

import (
	"errors"

	"github.com/turtacn/QuantaID/internal/domain/webhook"
	"gorm.io/gorm"
)

type WebhookRepository struct {
	db *gorm.DB
}

func NewWebhookRepository(db *gorm.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

func (r *WebhookRepository) Create(subscription *webhook.Subscription) error {
	return r.db.Create(subscription).Error
}

func (r *WebhookRepository) GetByID(id string) (*webhook.Subscription, error) {
	var sub webhook.Subscription
	if err := r.db.First(&sub, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, webhook.ErrSubscriptionNotFound
		}
		return nil, err
	}
	return &sub, nil
}

func (r *WebhookRepository) List() ([]*webhook.Subscription, error) {
	var subs []*webhook.Subscription
	if err := r.db.Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}

func (r *WebhookRepository) Update(subscription *webhook.Subscription) error {
	return r.db.Save(subscription).Error
}

func (r *WebhookRepository) Delete(id string) error {
	return r.db.Delete(&webhook.Subscription{}, "id = ?", id).Error
}

func (r *WebhookRepository) FindByEventType(eventType string) ([]*webhook.Subscription, error) {
	var subs []*webhook.Subscription
	// Using Postgres array operator to check if eventType is in Events array
	if err := r.db.Where("? = ANY(events)", eventType).Where("active = ?", true).Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}

package webhook

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/internal/domain/webhook"
)

type Service struct {
	repo webhook.Repository
}

func NewService(repo webhook.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateSubscription(url string, events []string) (*webhook.Subscription, error) {
	secret, err := generateSecret()
	if err != nil {
		return nil, err
	}

	sub := &webhook.Subscription{
		ID:        uuid.New().String(),
		URL:       url,
		Secret:    secret,
		Events:    events,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *Service) ListSubscriptions() ([]*webhook.Subscription, error) {
	return s.repo.List()
}

func (s *Service) GetSubscription(id string) (*webhook.Subscription, error) {
	return s.repo.GetByID(id)
}

func (s *Service) DeleteSubscription(id string) error {
	return s.repo.Delete(id)
}

func (s *Service) RotateSecret(id string) (*webhook.Subscription, error) {
	sub, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	newSecret, err := generateSecret()
	if err != nil {
		return nil, err
	}

	sub.Secret = newSecret
	sub.UpdatedAt = time.Now()

	if err := s.repo.Update(sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func generateSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

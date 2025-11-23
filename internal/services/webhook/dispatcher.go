package webhook

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/internal/domain/webhook"
	"github.com/turtacn/QuantaID/internal/worker"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

type Dispatcher struct {
	repo   webhook.Repository
	worker *worker.WebhookSender
	logger utils.Logger
}

func NewDispatcher(repo webhook.Repository, worker *worker.WebhookSender, logger utils.Logger) *Dispatcher {
	return &Dispatcher{
		repo:   repo,
		worker: worker,
		logger: logger,
	}
}

// Dispatch finds subscriptions matching the eventType and enqueues delivery tasks.
func (d *Dispatcher) Dispatch(ctx context.Context, eventType string, payload interface{}) {
	subs, err := d.repo.FindByEventType(eventType)
	if err != nil {
		d.logger.Error(ctx, "Failed to find subscriptions for event", zap.String("event_type", eventType), zap.Error(err))
		return
	}

	if len(subs) == 0 {
		return
	}

	eventID := uuid.New().String()
	d.logger.Info(ctx, "Dispatching webhook event", zap.String("event_type", eventType), zap.Int("subscribers", len(subs)))

	for _, sub := range subs {
		task := webhook.DeliveryTask{
			SubscriptionID: sub.ID,
			EventID:        eventID,
			EventType:      eventType,
			Payload:        payload,
			Attempt:        0,
			NextRetry:      time.Now(),
		}
		d.worker.Enqueue(task)
	}
}

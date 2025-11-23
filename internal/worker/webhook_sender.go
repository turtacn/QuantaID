package worker

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/turtacn/QuantaID/internal/domain/webhook"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

type WebhookSender struct {
	client     *http.Client
	queue      chan webhook.DeliveryTask
	repo       webhook.Repository
	logger     utils.Logger
	maxRetries int
}

func NewWebhookSender(repo webhook.Repository, logger utils.Logger, queueSize int) *WebhookSender {
	return &WebhookSender{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		queue:      make(chan webhook.DeliveryTask, queueSize),
		repo:       repo,
		logger:     logger,
		maxRetries: 3,
	}
}

func (w *WebhookSender) Start(ctx context.Context) {
	w.logger.Info(ctx, "Starting Webhook Sender Worker")
	go func() {
		for {
			select {
			case <-ctx.Done():
				w.logger.Info(ctx, "Stopping Webhook Sender Worker")
				return
			case task := <-w.queue:
				// Use a detached context for processing to avoid cancellation mid-process if possible,
				// but respect shutdown signals if critical.
				// For retries, we might want to block or spawn a goroutine.
				// Given the "asynchronous" requirement, spawning a goroutine per task or having a worker pool is better.
				// For simplicity here, we process sequentially in this worker routine,
				// but "exponential backoff" implies waiting. We shouldn't block the main loop.
				// So we should spawn a handler.
				go w.process(context.Background(), task)
			}
		}
	}()
}

func (w *WebhookSender) Enqueue(task webhook.DeliveryTask) {
	select {
	case w.queue <- task:
	default:
		w.logger.Warn(context.Background(), "Webhook queue full, dropping task", zap.String("subscription_id", task.SubscriptionID))
	}
}

func (w *WebhookSender) process(ctx context.Context, task webhook.DeliveryTask) {
	sub, err := w.repo.GetByID(task.SubscriptionID)
	if err != nil {
		w.logger.Error(ctx, "Failed to get subscription for webhook delivery", zap.Error(err), zap.String("subscription_id", task.SubscriptionID))
		return
	}

	payloadBytes, err := json.Marshal(task.Payload)
	if err != nil {
		w.logger.Error(ctx, "Failed to marshal webhook payload", zap.Error(err))
		return
	}
	payloadStr := string(payloadBytes)

	// Retry loop
	for attempt := 0; attempt <= w.maxRetries; attempt++ {
		// If this is a retry, wait before sending
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second // 2s, 4s, 8s... or prompt says 5s, 30s...
			// Prompt: 5s, 30s, 1m, 5m. Let's approximate or use a simple exponential for now.
			// Let's stick to simple exponential: 2^attempt * 1s -> 1s, 2s, 4s.
			// Prompt suggested: 5s, 30s, 1m.
			switch attempt {
			case 1:
				backoff = 5 * time.Second
			case 2:
				backoff = 30 * time.Second
			case 3:
				backoff = 1 * time.Minute
			default:
				backoff = 5 * time.Minute
			}

			w.logger.Info(ctx, "Retrying webhook delivery", zap.String("subscription_id", sub.ID), zap.Int("attempt", attempt), zap.Duration("backoff", backoff))
			time.Sleep(backoff)
		}

		ts := time.Now().Unix()
		signature := signPayload(sub.Secret, payloadStr, ts)

		req, err := http.NewRequestWithContext(ctx, "POST", sub.URL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			w.logger.Error(ctx, "Failed to create webhook request", zap.Error(err))
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-QuantaID-Signature", signature)
		req.Header.Set("X-QuantaID-Timestamp", fmt.Sprintf("%d", ts))
		req.Header.Set("X-QuantaID-Event-ID", task.EventID)
		req.Header.Set("X-QuantaID-Event-Type", task.EventType)

		resp, err := w.client.Do(req)
		if err != nil {
			w.logger.Warn(ctx, "Webhook delivery failed", zap.Error(err), zap.Int("attempt", attempt))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			w.logger.Info(ctx, "Webhook delivered successfully", zap.String("subscription_id", sub.ID), zap.Int("status", resp.StatusCode))
			return
		}

		w.logger.Warn(ctx, "Webhook delivery returned non-2xx status", zap.Int("status", resp.StatusCode), zap.Int("attempt", attempt))
	}

	w.logger.Error(ctx, "Webhook delivery failed after max retries", zap.String("subscription_id", sub.ID))
}

func signPayload(secret, body string, ts int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d.%s", ts, body)))
	return hex.EncodeToString(mac.Sum(nil))
}

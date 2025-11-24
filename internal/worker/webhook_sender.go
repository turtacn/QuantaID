package worker

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
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
				// Process async to avoid blocking queue consumption
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

	// If NextRetry is in the future, wait
	if time.Now().Before(task.NextRetry) {
		time.Sleep(time.Until(task.NextRetry))
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
	var statusCode int
	if resp != nil {
		statusCode = resp.StatusCode
		resp.Body.Close()
	}

	if err != nil || statusCode < 200 || statusCode >= 300 {
		w.logger.Warn(ctx, "Webhook delivery failed", zap.Error(err), zap.Int("status", statusCode), zap.Int("attempt", task.Attempt))

		// Retry Logic
		if task.Attempt < w.maxRetries {
			task.Attempt++
			// Exponential Backoff: 2^attempt * 1s
			backoff := time.Duration(math.Pow(2, float64(task.Attempt))) * time.Second
			task.NextRetry = time.Now().Add(backoff)
			w.logger.Info(ctx, "Scheduling webhook retry", zap.Int("attempt", task.Attempt), zap.Duration("backoff", backoff))

			// Re-enqueue or wait.
			// To avoid blocking *this* goroutine (which is detached per task), waiting here is acceptable for small scale.
			// But for better resource usage, we should re-enqueue with a delay mechanism.
			// Since we don't have a delayed queue, we will sleep here.
			time.Sleep(backoff)
			w.process(ctx, task) // Recursive retry
		} else {
			// DLQ Logic: Log to error or DB
			w.logger.Error(ctx, "Webhook delivery permanently failed", zap.String("subscription_id", sub.ID))
			// TODO: Implement actual DLQ storage
		}
		return
	}

	w.logger.Info(ctx, "Webhook delivered successfully", zap.String("subscription_id", sub.ID), zap.Int("status", statusCode))
}

func signPayload(secret, body string, ts int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d.%s", ts, body)))
	return hex.EncodeToString(mac.Sum(nil))
}

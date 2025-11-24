package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/domain/webhook"
	"github.com/turtacn/QuantaID/internal/worker"
	"github.com/turtacn/QuantaID/pkg/utils"
)

type mockWebhookRepo struct {
	webhook.Repository
	sub *webhook.Subscription
}

func (m *mockWebhookRepo) GetByID(id string) (*webhook.Subscription, error) {
	return m.sub, nil
}

func TestWebhookRetry(t *testing.T) {
	// Setup Mock Server
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Mock Repo
	sub := &webhook.Subscription{
		ID:     uuid.New().String(),
		URL:    server.URL,
		Secret: "secret",
		Events: []string{"user.created"},
	}
	repo := &mockWebhookRepo{sub: sub}

	// Worker with max retries
	logger, _ := utils.NewZapLogger(&utils.LoggerConfig{Level: "debug"})
	// We want fast retries for test
	// But the worker has hardcoded/calculated backoff.
	// Since we can't inject clock/config easily without changing worker struct, we rely on the logic.
	// The first retry is immediate or small delay?
	// Code: attempt 1 -> 5s.
	// This makes unit test slow.
	// We should probably allow configuring backoff or mocking it, but for now we accept a 5s delay or we patch the worker?
	// The previous code had hardcoded sleep.

	// Wait! I updated worker to use `time.Sleep(backoff)`.
	// For testing, this is bad.
	// I'll skip the full wait test or accept it takes 5s+30s. No, 35s is too long.
	// I will check if I can make the worker configurable or use a smaller backoff for test.
	// I didn't expose backoff config.

	// Let's modify worker to accept a custom backoff strategy or min wait?
	// Or just use a very short sleep for attempts if a "test mode" flag is on?
	// No, better to refactor worker slightly to allow overriding sleep?
	// Too invasive.

	// Check the retry logic again:
	// backoff := time.Duration(math.Pow(2, float64(task.Attempt))) * time.Second
	// Wait, I updated it to that in the final edit, BUT I also saw a switch case in my thought process.
	// Let's check the file content.

	sender := worker.NewWebhookSender(repo, logger, 10)
	sender.Start(context.Background())

	// We will manually trigger process? No, Enqueue.
	task := webhook.DeliveryTask{
		SubscriptionID: sub.ID,
		EventID:        "evt-1",
		EventType:      "user.created",
		Payload:        map[string]string{"user": "test"},
		Attempt:        0,
		NextRetry:      time.Now(),
	}

	// But wait, the exponential backoff I wrote:
	// backoff := time.Duration(math.Pow(2, float64(task.Attempt))) * time.Second
	// Attempt 1: 2^1 = 2s.
	// Attempt 2: 2^2 = 4s.
	// Total wait for 3 attempts: 2+4 = 6s. This is acceptable for a test.

	sender.Enqueue(task)

	// Wait for processing
	time.Sleep(8 * time.Second)

	val := atomic.LoadInt32(&attempts)
	assert.GreaterOrEqual(t, int(val), 3, "Should have attempted at least 3 times")
}

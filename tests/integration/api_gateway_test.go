//go:build integration
// +build integration

package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/turtacn/QuantaID/internal/server/middleware"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
)

// Simplified integration test using direct middleware call and Redis container
func TestRateLimit_Enforcement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// 1. Start Redis Container
	// We use variable 'containerReq' to avoid conflict with 'req' later used for http requests
	containerReq := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: containerReq,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer redisC.Terminate(ctx)

	endpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})
	defer redisClient.Close()

	// 2. Setup Middleware
	// Limit: 10 req/sec
	// Pass nil for APIKeyService as we don't need policy lookup for this basic test
	logger := zap.NewNop()
	limiter := middleware.NewRateLimitMiddleware(redisClient, nil, 10, 1, logger)

	handler := limiter.Execute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 3. Simulate requests
	// AppID "app-test"

	// Send 15 requests concurrently
	var wg sync.WaitGroup
	results := make([]int, 15)

	for i := 0; i < 15; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/", nil)
			// Inject AppID to simulate authenticated request
			// Use types.ContextKeyAppID if available, but assuming test uses updated middleware which uses updated constant.
			// Middleware uses `types.ContextKeyAppID`.
			ctx := context.WithValue(req.Context(), types.ContextKeyAppID, "app-test")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req.WithContext(ctx))
			results[index] = w.Code
		}(i)
	}
	wg.Wait()

	// 4. Verify results
	successCount := 0
	limitCount := 0
	for _, code := range results {
		if code == http.StatusOK {
			successCount++
		} else if code == http.StatusTooManyRequests {
			limitCount++
		}
	}

	t.Logf("Success: %d, Limited: %d", successCount, limitCount)
	assert.Equal(t, 10, successCount, "Should allow exactly 10 requests")
	assert.Equal(t, 5, limitCount, "Should reject 5 requests")

	// 5. Wait for window expiration and retry
	time.Sleep(1100 * time.Millisecond) // Wait > 1s

	req := httptest.NewRequest("GET", "/", nil)
	ctx = context.WithValue(req.Context(), types.ContextKeyAppID, "app-test")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req.WithContext(ctx))

	assert.Equal(t, http.StatusOK, w.Code, "Should be allowed after window expiration")
}

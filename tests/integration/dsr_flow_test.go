//go:build integration
package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/server/http"
	"github.com/turtacn/QuantaID/internal/services/privacy"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	http2 "net/http"
	"os"
	"strconv"
)

func createTestServer(t *testing.T) *http.Server {
	logger, _ := utils.NewZapLogger(&utils.LoggerConfig{Level: "error"})
	configManager, err := utils.NewConfigManager("../../configs", "server", "yaml", logger)
	assert.NoError(t, err)
	var appCfg utils.Config
	err = configManager.Unmarshal(&appCfg)
	assert.NoError(t, err)

	pgPort, _ := strconv.Atoi(os.Getenv("QID_POSTGRES_PORT"))
	redisPort, _ := strconv.Atoi(os.Getenv("QID_REDIS_PORT"))

	appCfg.Postgres.Host = os.Getenv("QID_POSTGRES_HOST")
	appCfg.Postgres.Port = pgPort
	appCfg.Postgres.User = os.Getenv("QID_POSTGRES_USER")
	appCfg.Postgres.Password = os.Getenv("QID_POSTGRES_PASSWORD")
	appCfg.Postgres.DbName = os.Getenv("QID_POSTGRES_DBNAME")
	appCfg.Redis.Host = os.Getenv("QID_REDIS_HOST")
	appCfg.Redis.Port = redisPort
	appCfg.Storage.Mode = "postgres"

	cryptoManager := utils.NewCryptoManager("test-secret")

	httpServer, err := http.NewServerWithConfig(http.Config{
		Address:      ":8081",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}, &appCfg, logger, cryptoManager)
	assert.NoError(t, err)

	return httpServer
}

func Test_Export_Flow(t *testing.T) {
	server := createTestServer(t)
	go server.Start() // Using Start() instead of ListenAndServe() on internal server struct
	defer func() {
		// Stop graceful shutdown logic or just close raw listener?
		// internal server has Stop(ctx)
		server.Stop(context.Background())
	}()

	// 1. Create a user
	user := createTestUser(t, server, "export_user", "export@example.com", "password")

	// 2. Create a JWT for the user
	token := createTestJWT(t, user)

	// 3. Make a request to the export endpoint
	req, err := http2.NewRequest("POST", "http://localhost:8081/api/v1/privacy/export", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http2.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// 4. Assert the results
	assert.Equal(t, http2.StatusOK, resp.StatusCode)

	var exportData privacy.ExportData
	err = json.NewDecoder(resp.Body).Decode(&exportData)
	assert.NoError(t, err)

	assert.Equal(t, user.ID, exportData.User.ID)
}

func createTestUser(t *testing.T, server *http.Server, username, email, password string) *types.User {
	// a bit of a hack to get the service
	s := server.Services.IdentityService
	user, err := s.CreateUser(context.Background(), username, email, password)
	assert.NoError(t, err)
	return user
}

func createTestJWT(t *testing.T, user *types.User) string {
	cryptoManager := utils.NewCryptoManager("test-secret")
	token, err := cryptoManager.GenerateJWT(user.ID, time.Hour, "openid")
	assert.NoError(t, err)
	return token
}

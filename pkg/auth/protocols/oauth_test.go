package protocols

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"testing"
)

type MockApplicationRepository struct {
	mock.Mock
}

func (m *MockApplicationRepository) GetApplicationByClientID(ctx context.Context, clientID string) (*types.Application, error) {
	args := m.Called(ctx, clientID)
	return args.Get(0).(*types.Application), args.Error(1)
}

func (m *MockApplicationRepository) CreateApplication(ctx context.Context, app *types.Application) error {
	return nil
}

func (m *MockApplicationRepository) GetApplicationByID(ctx context.Context, id string) (*types.Application, error) {
	return nil, nil
}

func (m *MockApplicationRepository) GetApplicationByName(ctx context.Context, name string) (*types.Application, error) {
	return nil, nil
}

func (m *MockApplicationRepository) UpdateApplication(ctx context.Context, app *types.Application) error {
	return nil
}

func (m *MockApplicationRepository) DeleteApplication(ctx context.Context, id string) error {
	return nil
}

func (m *MockApplicationRepository) ListApplications(ctx context.Context, pq types.PaginationQuery) ([]*types.Application, error) {
	return nil, nil
}

func TestOAuthAdapter_HandleAuthRequest(t *testing.T) {
	logger := utils.NewZapLoggerWrapper(zap.NewNop())
	mockAppRepo := new(MockApplicationRepository)
	mockRedis := new(redis.MockRedisClient)
	adapter := &OAuthAdapter{
		logger:  logger,
		appRepo: mockAppRepo,
		redis:   mockRedis,
	}

	app := &types.Application{
		ClientType: types.ClientTypePublic,
		ProtocolConfig: map[string]interface{}{
			"redirect_uris": []string{"http://localhost:3000/callback"},
		},
	}
	mockAppRepo.On("GetApplicationByClientID", mock.Anything, "test_client_id").Return(app, nil)
	mockRedis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	req := &types.AuthRequest{
		Credentials: map[string]string{
			"response_type":         "code",
			"client_id":             "test_client_id",
			"redirect_uri":          "http://localhost:3000/callback",
			"scope":                 "openid",
			"state":                 "test_state",
			"code_challenge":        "test_code_challenge",
			"code_challenge_method": "S256",
		},
	}

	resp, err := adapter.HandleAuthRequest(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Code)
}

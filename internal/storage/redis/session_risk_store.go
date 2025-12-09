package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	sessionRiskKeyPrefix = "session:risk:"
	sessionRiskTTL       = 30 * time.Minute
)

// SessionRiskData holds the risk information for a session.
type SessionRiskData struct {
	SessionID      string    `json:"session_id"`
	RiskScore      int       `json:"risk_score"`
	RiskLevel      string    `json:"risk_level"`
	LastEvaluated  time.Time `json:"last_evaluated"`
	Signals        []string  `json:"signals"`       // Active risk signals
	ActionsTaken   []string  `json:"actions_taken"` // Actions already executed
	NextEvaluation time.Time `json:"next_evaluation"`
}

// SessionRiskStore handles storage of session risk data in Redis.
type SessionRiskStore struct {
	client RedisClientInterface
}

// NewSessionRiskStore creates a new SessionRiskStore.
func NewSessionRiskStore(client RedisClientInterface) *SessionRiskStore {
	return &SessionRiskStore{
		client: client,
	}
}

// Set stores the risk data for a session.
func (s *SessionRiskStore) Set(ctx context.Context, sessionID string, data SessionRiskData) error {
	key := sessionRiskKeyPrefix + sessionID
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, jsonData, sessionRiskTTL)
}

// Get retrieves the risk data for a session.
func (s *SessionRiskStore) Get(ctx context.Context, sessionID string) (*SessionRiskData, error) {
	key := sessionRiskKeyPrefix + sessionID
	val, err := s.client.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var data SessionRiskData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// UpdateScore updates the risk score and level for a session.
func (s *SessionRiskStore) UpdateScore(ctx context.Context, sessionID string, score int, level string) error {
	data, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}
	if data == nil {
		data = &SessionRiskData{SessionID: sessionID}
	}

	data.RiskScore = score
	data.RiskLevel = level
	data.LastEvaluated = time.Now()

	return s.Set(ctx, sessionID, *data)
}

// AddSignal adds a risk signal to the session risk data.
func (s *SessionRiskStore) AddSignal(ctx context.Context, sessionID, signal string) error {
	data, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}
	if data == nil {
		data = &SessionRiskData{SessionID: sessionID}
	}

	exists := false
	for _, sig := range data.Signals {
		if sig == signal {
			exists = true
			break
		}
	}
	if !exists {
		data.Signals = append(data.Signals, signal)
	}

	return s.Set(ctx, sessionID, *data)
}

// ClearSignals clears all risk signals for a session.
func (s *SessionRiskStore) ClearSignals(ctx context.Context, sessionID string) error {
	data, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}
	if data == nil {
		return nil
	}

	data.Signals = []string{}
	return s.Set(ctx, sessionID, *data)
}

// Delete removes the risk data for a session.
func (s *SessionRiskStore) Delete(ctx context.Context, sessionID string) error {
	key := sessionRiskKeyPrefix + sessionID
	return s.client.Del(ctx, key)
}

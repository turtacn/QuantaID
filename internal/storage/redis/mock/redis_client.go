package mock

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

type RedisClient struct {
	mock.Mock
}

func (m *RedisClient) Client() *redis.Client {
	args := m.Called()
	return args.Get(0).(*redis.Client)
}

func (m *RedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *RedisClient) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *RedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *RedisClient) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *RedisClient) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *RedisClient) SCard(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]string), args.Error(1)
}

func (m *RedisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *RedisClient) ZCard(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *RedisClient) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).(int64), args.Error(1)
}

func (m *RedisClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

func (m *RedisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).([]string), args.Error(1)
}

func (m *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

func (m *RedisClient) SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *RedisClient) SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	args := m.Called(ctx, key, member)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *RedisClient) GeoAdd(ctx context.Context, key string, locations ...*redis.GeoLocation) (int64, error) {
	args := m.Called(ctx, key, locations)
	return args.Get(0).(int64), args.Error(1)
}

func (m *RedisClient) GeoPos(ctx context.Context, key string, members ...string) ([]*redis.GeoPos, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).([]*redis.GeoPos), args.Error(1)
}

func (m *RedisClient) HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *RedisClient) HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.MapStringStringCmd)
}

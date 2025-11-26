package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/turtacn/QuantaID/internal/security/automator"
	"github.com/turtacn/QuantaID/internal/storage/redis"
)

// BlockIPAction implements the SecurityAction interface to block an IP address.
type BlockIPAction struct {
	redisClient redis.RedisClientInterface
}

// NewBlockIPAction creates a new instance of BlockIPAction.
func NewBlockIPAction(redisClient redis.RedisClientInterface) *BlockIPAction {
	return &BlockIPAction{redisClient: redisClient}
}

// ID returns the unique identifier for the action.
func (a *BlockIPAction) ID() string {
	return "block_ip"
}

// Execute adds the IP address from the input to a Redis blacklist with a 1-hour TTL.
func (a *BlockIPAction) Execute(ctx context.Context, input automator.ActionInput) error {
	if input.IP == "" {
		return fmt.Errorf("IP address is required for BlockIPAction")
	}
	key := fmt.Sprintf("security:blacklist:ip:%s", input.IP)
	// Block for 1 hour
	return a.redisClient.SetEx(ctx, key, "1", time.Hour).Err()
}

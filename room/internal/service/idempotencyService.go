package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	messageIDKeyPrefix       = "message_id:"
	messageStatusUncompleted = "uncompleted"
	messageStatusCompleted   = "completed"
	messageIDTTL             = time.Hour
)

var (
	ErrMessageIDAlreadyCompleted = errors.New("message ID is already completed")
	ErrMessageIDNotUncompleted   = errors.New("message ID is missing or is not uncompleted")
)

// IdempotencyService validates the shared Gateway/Room message state in Redis
// and marks a message completed only after its Room operation succeeds.
type IdempotencyService struct {
	store MessageStateStore
}

// MessageStateStore is the Redis-backed state contract used by the service.
type MessageStateStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
}

type redisMessageStateStore struct{ client *redis.Client }

func (s redisMessageStateStore) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

func (s redisMessageStateStore) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return s.client.Set(ctx, key, value, expiration).Err()
}

func NewIdempotencyService(client *redis.Client) *IdempotencyService {
	return &IdempotencyService{store: redisMessageStateStore{client: client}}
}

func NewIdempotencyServiceWithStore(store MessageStateStore) *IdempotencyService {
	return &IdempotencyService{store: store}
}

func (s *IdempotencyService) Run(ctx context.Context, messageID string, operation func() error) error {
	if strings.TrimSpace(messageID) == "" {
		return status.Error(codes.FailedPrecondition, "message_id is required")
	}

	state, err := s.store.Get(ctx, messageIDKeyPrefix+messageID)
	if err == redis.Nil {
		return status.Error(codes.FailedPrecondition, ErrMessageIDNotUncompleted.Error())
	}
	if err != nil {
		return status.Errorf(codes.Unavailable, "read message ID state from Redis: %v", err)
	}
	if state == messageStatusCompleted {
		return status.Error(codes.AlreadyExists, ErrMessageIDAlreadyCompleted.Error())
	}
	if state != messageStatusUncompleted {
		return status.Error(codes.FailedPrecondition, ErrMessageIDNotUncompleted.Error())
	}

	if err := operation(); err != nil {
		return err
	}
	if err := s.store.Set(ctx, messageIDKeyPrefix+messageID, messageStatusCompleted, messageIDTTL); err != nil {
		return status.Errorf(codes.Unavailable, "mark message ID completed in Redis: %v", err)
	}
	return nil
}

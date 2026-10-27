package retry

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Policy controls bounded retries for safe, repeatable downstream RPCs.
// MaxAttempts includes the initial request.
type Policy struct {
	MaxAttempts int
	Interval    time.Duration
}

func DefaultPolicy() Policy {
	return Policy{MaxAttempts: 3, Interval: 200 * time.Millisecond}
}

func PolicyFromEnv() (Policy, error) {
	policy := DefaultPolicy()
	if value := os.Getenv("GATEWAY_GRPC_RETRY_MAX_ATTEMPTS"); value != "" {
		attempts, err := strconv.Atoi(value)
		if err != nil || attempts < 1 {
			return Policy{}, fmt.Errorf("invalid GATEWAY_GRPC_RETRY_MAX_ATTEMPTS %q: expected a positive integer", value)
		}
		policy.MaxAttempts = attempts
	}
	if value := os.Getenv("GATEWAY_GRPC_RETRY_INTERVAL"); value != "" {
		interval, err := time.ParseDuration(value)
		if err != nil || interval <= 0 {
			return Policy{}, fmt.Errorf("invalid GATEWAY_GRPC_RETRY_INTERVAL %q: expected a positive duration", value)
		}
		policy.Interval = interval
	}
	return policy, nil
}

// Do repeats only UNAVAILABLE calls and only while the caller's context has
// time remaining. Callers must use this for operations safe to repeat.
func Do[T any](ctx context.Context, policy Policy, call func(context.Context) (T, error)) (T, error) {
	var zero T
	if policy.MaxAttempts < 1 || policy.Interval <= 0 {
		return zero, fmt.Errorf("invalid gRPC retry policy")
	}
	for attempt := 1; ; attempt++ {
		result, err := call(ctx)
		if err == nil || status.Code(err) != codes.Unavailable || attempt >= policy.MaxAttempts {
			return result, err
		}

		timer := time.NewTimer(policy.Interval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return zero, ctx.Err()
		case <-timer.C:
		}
	}
}

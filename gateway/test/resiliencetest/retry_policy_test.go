package resiliencetest

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"gateway/internal/client/retry"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDoRetriesUnavailableUntilSuccess(t *testing.T) {
	var attempts atomic.Int32
	result, err := retry.Do(context.Background(), retry.Policy{MaxAttempts: 3, Interval: time.Millisecond}, func(context.Context) (string, error) {
		if attempts.Add(1) < 3 {
			return "", status.Error(codes.Unavailable, "temporary")
		}
		return "ok", nil
	})
	if err != nil || result != "ok" {
		t.Fatalf("Do() = (%q, %v), want (ok, nil)", result, err)
	}
	if got := attempts.Load(); got != 3 {
		t.Fatalf("attempts = %d, want 3", got)
	}
}

func TestDoDoesNotRetryOtherStatusCodes(t *testing.T) {
	var attempts atomic.Int32
	_, err := retry.Do(context.Background(), retry.Policy{MaxAttempts: 3, Interval: time.Millisecond}, func(context.Context) (string, error) {
		attempts.Add(1)
		return "", status.Error(codes.DeadlineExceeded, "deadline")
	})
	if status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("error = %v, want DeadlineExceeded", err)
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("attempts = %d, want 1", got)
	}
}

func TestDoStopsWaitingWhenContextExpires(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	var attempts atomic.Int32
	started := time.Now()
	_, err := retry.Do(ctx, retry.Policy{MaxAttempts: 3, Interval: time.Second}, func(context.Context) (string, error) {
		attempts.Add(1)
		return "", status.Error(codes.Unavailable, "temporary")
	})
	if err != context.DeadlineExceeded {
		t.Fatalf("error = %v, want context deadline exceeded", err)
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("attempts = %d, want 1", got)
	}
	if elapsed := time.Since(started); elapsed > 250*time.Millisecond {
		t.Fatalf("retry wait ignored context deadline: %v", elapsed)
	}
}

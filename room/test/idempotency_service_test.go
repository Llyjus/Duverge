package service

import (
	"context"
	"errors"
	"testing"

	idempotency "room/internal/service"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeMessageStateStore struct {
	state  string
	getErr error
	setErr error
}

func (s *fakeMessageStateStore) Get(context.Context, string) (string, error) {
	return s.state, s.getErr
}

func (s *fakeMessageStateStore) Set(_ context.Context, _ string, value string) error {
	if s.setErr == nil {
		s.state = value
	}
	return s.setErr
}

func TestIdempotencyServiceRunMarksSuccessfulOperationCompleted(t *testing.T) {
	store := &fakeMessageStateStore{state: "uncompleted"}
	service := idempotency.NewIdempotencyServiceWithStore(store)
	called := false

	err := service.Run(context.Background(), "id-1", func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !called || store.state != "completed" {
		t.Fatalf("operation called=%v, Redis status=%q; want called and completed", called, store.state)
	}
}

func TestIdempotencyServiceRunLeavesMessageUncompletedOnOperationError(t *testing.T) {
	store := &fakeMessageStateStore{state: "uncompleted"}
	service := idempotency.NewIdempotencyServiceWithStore(store)
	wantErr := errors.New("operation failed")

	err := service.Run(context.Background(), "id-1", func() error { return wantErr })
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
	if store.state != "uncompleted" {
		t.Fatalf("Redis status = %q, want uncompleted", store.state)
	}
}

func TestIdempotencyServiceRunRejectsInvalidMessageStates(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		state    string
		getErr   error
		wantCode codes.Code
	}{
		{name: "empty ID", id: "", state: "uncompleted", wantCode: codes.FailedPrecondition},
		{name: "missing ID", id: "id-1", getErr: redis.Nil, wantCode: codes.FailedPrecondition},
		{name: "unknown state", id: "id-1", state: "pending", wantCode: codes.FailedPrecondition},
		{name: "already completed", id: "id-1", state: "completed", wantCode: codes.AlreadyExists},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeMessageStateStore{state: tt.state, getErr: tt.getErr}
			called := false
			err := idempotency.NewIdempotencyServiceWithStore(store).Run(context.Background(), tt.id, func() error {
				called = true
				return nil
			})
			if status.Code(err) != tt.wantCode {
				t.Fatalf("status code = %s, want %s (err=%v)", status.Code(err), tt.wantCode, err)
			}
			if called {
				t.Fatal("operation ran for invalid message state")
			}
		})
	}
}

func TestIdempotencyServiceRunReturnsRedisCompletionError(t *testing.T) {
	storeErr := errors.New("Redis unavailable")
	store := &fakeMessageStateStore{state: "uncompleted", setErr: storeErr}
	err := idempotency.NewIdempotencyServiceWithStore(store).Run(context.Background(), "id-1", func() error { return nil })
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("Run() code = %s, want %s (err=%v)", status.Code(err), codes.Unavailable, err)
	}
	if store.state != "uncompleted" {
		t.Fatalf("Redis status = %q, want uncompleted after failed write", store.state)
	}
}

package resiliencetest

import (
	"context"
	"errors"
	"testing"

	internalerrors "gateway/internal/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIsRequestTimeout(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "context deadline", err: context.DeadlineExceeded, want: true},
		{name: "wrapped context deadline", err: errors.Join(errors.New("request failed"), context.DeadlineExceeded), want: true},
		{name: "gRPC deadline", err: status.Error(codes.DeadlineExceeded, "deadline exceeded"), want: true},
		{name: "other gRPC status", err: status.Error(codes.Unavailable, "unavailable"), want: false},
		{name: "ordinary error", err: errors.New("failure"), want: false},
		{name: "nil", err: nil, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := internalerrors.IsRequestTimeout(test.err); got != test.want {
				t.Fatalf("IsRequestTimeout(%v) = %v, want %v", test.err, got, test.want)
			}
		})
	}
}

package errors

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrRedisUnavailable marks failures to reach or use Redis. Callers can map it
// to a service-unavailable response without treating infrastructure failures
// as invalid user credentials.
var ErrRedisUnavailable = errors.New("Redis is unavailable")

// ErrUserDataUnavailable marks a failed UserData gRPC dependency call.
var ErrUserDataUnavailable = errors.New("UserData service is unavailable")

// IsRequestTimeout reports whether err represents a context or gRPC deadline.
func IsRequestTimeout(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || status.Code(err) == codes.DeadlineExceeded
}

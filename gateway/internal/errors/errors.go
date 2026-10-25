package errors

import "errors"

// ErrRedisUnavailable marks failures to reach or use Redis. Callers can map it
// to a service-unavailable response without treating infrastructure failures
// as invalid user credentials.
var ErrRedisUnavailable = errors.New("Redis is unavailable")

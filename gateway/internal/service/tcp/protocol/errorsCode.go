package protocol

const (
	ErrCodeInvalidMessage  = "INVALID_MESSAGE"
	ErrCodeUnauthenticated = "UNAUTHENTICATED"
	ErrCodeForbidden       = "FORBIDDEN"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeInternal        = "INTERNAL_ERROR"

	ErrCodeVersionMismatch = "VERSION_MISMATCH"
	ErrCodeServerClosed    = "SERVER_CLOSED_ERROR"
)

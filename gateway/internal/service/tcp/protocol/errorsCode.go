package protocol

const (
	ErrCodeInvalidMessage  = "INVALID_MESSAGE"
	ErrCodeUnauthenticated = "UNAUTHENTICATED"
	ErrCodeForbidden       = "FORBIDDEN"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeInternal        = "INTERNAL_ERROR"
	ErrCodeServerErr       = "SERVER_ERROR"
	ErrCodeServerClosedErr = "SERVER_CLOSED_ERROR"
)

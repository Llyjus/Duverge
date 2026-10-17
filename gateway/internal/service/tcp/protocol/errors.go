package protocol

import "errors"

var ErrInvalidMessage = errors.New("invalid message")
var ErrUnauthenticated = errors.New("unauthenticated connection")
var ErrForbidden = errors.New("forbidden action")
var ErrNotFound = errors.New("resource not found")
var ErrInternal = errors.New("internal server error")
var ErrAuthenticationFailed = errors.New("authentication failed")

var ErrVersionMismatch = errors.New("version mismatch")
var ErrServerClosed = errors.New("server closed the connection")

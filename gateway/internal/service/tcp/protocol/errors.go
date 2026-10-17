package protocol

import "errors"

var ErrInvalidMessage = errors.New("invalid chat message")
var ErrUnauthenticated = errors.New("unauthenticated connection")
var ErrForbidden = errors.New("forbidden action")
var ErrNotFound = errors.New("resource not found")
var ErrInternal = errors.New("internal server error")
var ServerErr = errors.New("server error")
var ServerClosedErr = errors.New("server closed the connection")

package errors

import "errors"

var ErrTooManyRequests = errors.New("too many requests")
var ErrInternalServer = errors.New("internal server error")

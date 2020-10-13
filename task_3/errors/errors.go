package errors

import "errors"

var (
	Interrupted = errors.New("process interrupted, please resume by resending the request")
	Shutdown    = errors.New("service is shutting down")
	Stopped     = errors.New("service is not running")
)

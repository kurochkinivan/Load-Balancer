package proxy

import "errors"

var (
	ErrNoBackendsAvailable = errors.New("there are no servers to process the request, try again later")
	ErrBackendRefusedConnection = errors.New("backend refused connection")
)
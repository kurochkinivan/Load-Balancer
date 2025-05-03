package proxy

import "errors"

var (
	ErrNoServicesAvailable = errors.New("there are no servers to process the request, try again later")
)
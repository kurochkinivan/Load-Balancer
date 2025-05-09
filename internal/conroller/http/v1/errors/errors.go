package httperror

import (
	"net/http"
)

// Rate Limiting errors
var (
	ErrRateLimitExceeded = New(nil, "rate limit exceeded", http.StatusTooManyRequests)
	ErrUnknownClient     = New(nil, "unknown client", http.StatusForbidden)
)

// Proxy errors
var (
	ErrNoBackendsAvailable = New(nil, "there are no servers to process the request, try again later", http.StatusServiceUnavailable)
)

func ErrDeserialize(err error) *HTTPError {
	return BadRequest(err, "failed to deserialize data")
}

func ErrSerialize(err error) *HTTPError {
	return InternalServerError(err, "failed to serialize data")
}

func BadRequest(err error, message string) *HTTPError {
	return New(err, message, http.StatusBadRequest)
}

func NotFound(err error, message string) *HTTPError {
	return New(err, message, http.StatusNotFound)
}

func Conflict(err error, message string) *HTTPError {
	return New(err, message, http.StatusConflict)
}

func InternalServerError(err error, message string) *HTTPError {
	return New(err, message, http.StatusInternalServerError)
}

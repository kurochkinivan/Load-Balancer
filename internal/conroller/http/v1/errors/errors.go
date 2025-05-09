package httperror

import (
	"net/http"
)

// Rate Limiting errors
var (
	// ErrRateLimitExceeded is returned when the client has exceeded the allowed rate limit.
	ErrRateLimitExceeded = New(nil, "rate limit exceeded", http.StatusTooManyRequests)

	// ErrUnknownClient is returned when the client is unknown.
	ErrUnknownClient = New(nil, "unknown client", http.StatusForbidden)
)

// Proxy errors
var (
	// ErrNoBackendsAvailable is returned when there are no servers to process the request.
	ErrNoBackendsAvailable = New(nil, "there are no servers to process the request, try again later", http.StatusServiceUnavailable)
)

// ErrDeserialize returns an HTTP error for deserialization errors.
func ErrDeserialize(err error) *HTTPError {
	return BadRequest(err, "failed to deserialize data")
}

// ErrSerialize returns an HTTP error for serialization errors.
func ErrSerialize(err error) *HTTPError {
	return InternalServerError(err, "failed to serialize data")
}

// BadRequest returns an HTTP error for bad requests.
func BadRequest(err error, message string) *HTTPError {
	return New(err, message, http.StatusBadRequest)
}

// NotFound returns an HTTP error for not found requests.
func NotFound(err error, message string) *HTTPError {
	return New(err, message, http.StatusNotFound)
}

// Conflict returns an HTTP error for conflicts.
func Conflict(err error, message string) *HTTPError {
	return New(err, message, http.StatusConflict)
}

// InternalServerError returns an HTTP error for internal server errors.
func InternalServerError(err error, message string) *HTTPError {
	return New(err, message, http.StatusInternalServerError)
}

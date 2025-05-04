package apperror

import (
	"net/http"
)

var (
	ErrRateLimitExceeded = New("rate limit exceeded", http.StatusTooManyRequests)
	ErrUnknownClient     = New("unknown client", http.StatusForbidden)
)

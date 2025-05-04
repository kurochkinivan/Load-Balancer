package middleware

import "net/http"

func RateLimitingMiddleware(next http.Handler) http.Handler {
	return nil
}

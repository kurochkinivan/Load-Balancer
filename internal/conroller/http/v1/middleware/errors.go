package middleware

import (
	"errors"
	"net/http"

	apperror "github.com/kurochkinivan/load_balancer/internal/lib/appError"
)

type errorHandler func(w http.ResponseWriter, r *http.Request) error

func ErrorMiddleware(next errorHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := next(w, r)

		if err != nil {
			var appErr *apperror.HTTPError
			if errors.As(err, &appErr) {
				w.WriteHeader(appErr.Code)
				w.Write(appErr.Marshal())
				return
			}

			appErr = apperror.New(err.Error(), http.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(appErr.Marshal())
		}
	})
}

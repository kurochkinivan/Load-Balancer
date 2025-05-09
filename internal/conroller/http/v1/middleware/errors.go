package middleware

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	httperror "github.com/kurochkinivan/load_balancer/internal/conroller/http/v1/errors"
)

// ErrorMiddleware catches errors from the next handler and returns an appropriate HTTP response.
// If the error is an instance of httperror.HTTPError, it writes the error as JSON and returns the status code.
// If the error is not an instance of httperror.HTTPError, it wraps it in an InternalServerError and writes the error as JSON.
func ErrorMiddleware(next AppHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := next(w, r)

		if err != nil {
			var httpError *httperror.HTTPError
			if errors.As(err, &httpError) {
				// If the error is an instance of httperror.HTTPError, write the error as JSON and return the status code.
				w.WriteHeader(httpError.Code)
				w.Write(httpError.Marshal())
				return
			}

			// If the error is not an instance of httperror.HTTPError, wrap it in an InternalServerError and write the error as JSON.
			httpError = httperror.InternalServerError(err, "")
			w.WriteHeader(httpError.Code)
			w.Write(httpError.Marshal())
		}
	})
}

// ErrorMiddlewareParams is similar to ErrorMiddleware, but it takes an httprouter.Params as an argument.
func ErrorMiddlewareParams(next ParamsAppHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		err := next(w, r, p)

		if err != nil {
			var httpError *httperror.HTTPError
			if errors.As(err, &httpError) {
				// If the error is an instance of httperror.HTTPError, write the error as JSON and return the status code.
				w.WriteHeader(httpError.Code)
				w.Write(httpError.Marshal())
				return
			}

			// If the error is not an instance of httperror.HTTPError, wrap it in an InternalServerError and write the error as JSON.
			httpError = httperror.InternalServerError(err, "")
			w.WriteHeader(httpError.Code)
			w.Write(httpError.Marshal())
		}
	}
}


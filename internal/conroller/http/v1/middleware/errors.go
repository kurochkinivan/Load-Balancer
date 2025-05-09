package middleware

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	httperror "github.com/kurochkinivan/load_balancer/internal/conroller/http/v1/errors"
)

func ErrorMiddleware(next AppHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := next(w, r)

		if err != nil {
			var httpError *httperror.HTTPError
			if errors.As(err, &httpError) {
				w.WriteHeader(httpError.Code)
				w.Write(httpError.Marshal())
				return
			}

			httpError = httperror.InternalServerError(err, "")
			w.WriteHeader(httpError.Code)
			w.Write(httpError.Marshal())
		}
	})
}

func ErrorMiddlewareParams(next ParamsAppHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		err := next(w, r, p)

		if err != nil {
			var httpError *httperror.HTTPError
			if errors.As(err, &httpError) {
				w.WriteHeader(httpError.Code)
				w.Write(httpError.Marshal())
				return
			}

			httpError = httperror.InternalServerError(err, "")
			w.WriteHeader(httpError.Code)
			w.Write(httpError.Marshal())
		}
	}
}

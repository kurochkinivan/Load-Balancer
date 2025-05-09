package middleware

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// ParamsAppHandler is a function that takes an httprouter.Handle, and returns an error.
type ParamsAppHandler func(http.ResponseWriter, *http.Request, httprouter.Params) error

// AppHandler is a function that takes an http.Handler, and returns an error.
type AppHandler func(w http.ResponseWriter, r *http.Request) error
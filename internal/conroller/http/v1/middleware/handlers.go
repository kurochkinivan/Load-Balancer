package middleware

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type ParamsAppHandler func(http.ResponseWriter, *http.Request, httprouter.Params) error

type AppHandler func(w http.ResponseWriter, r *http.Request) error
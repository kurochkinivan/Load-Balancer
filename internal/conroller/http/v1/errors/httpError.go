package httperror

import (
	"encoding/json"
	"fmt"
)

// HTTPError представляет ошибку с HTTP-статусом
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func New(err error, message string, status int) *HTTPError {
	if err != nil {
		message = fmt.Sprintf("%s: %v", message, err)
	}

	return &HTTPError{
		Message: message,
		Code:    status,
	}
}

func (e *HTTPError) Error() string {
	return e.Message
}

func (e *HTTPError) StatusCode() int {
	return e.Code
}

func (e *HTTPError) Marshal() []byte {
	data, _ := json.Marshal(e)
	return data
}

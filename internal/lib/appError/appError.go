package apperror

import (
	"encoding/json"
)

// HTTPError представляет ошибку с HTTP-статусом
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func New(message string, status int) *HTTPError {
	return &HTTPError{
		Message:  message,
		Code: status,
	}
}

func (e *HTTPError) Error() string {
	return e.Message
}

func (e *HTTPError) StatusCode() int {
	return e.Code
}

func (e *HTTPError) Marshal() []byte {
	data, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return data
}

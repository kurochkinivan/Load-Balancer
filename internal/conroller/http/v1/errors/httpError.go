// Package httperror provides a way to wrap errors with HTTP status codes.
//
// HTTPError type is used to represent errors with HTTP status codes. It
// implements the error interface and provides methods to get the status code and
// the error message.
package httperror

import (
	"encoding/json"
	"fmt"
)

// HTTPError represents an error with HTTP status code.
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// New creates a new HTTPError with the given error, message and status code.
// If the error is not nil, it appends the error message to the message.
func New(err error, message string, status int) *HTTPError {
	if err != nil {
		message = fmt.Sprintf("%s: %v", message, err)
	}

	return &HTTPError{
		Message: message,
		Code:    status,
	}
}

// Error returns the error message.
func (e *HTTPError) Error() string {
	return e.Message
}

// StatusCode returns the HTTP status code.
func (e *HTTPError) StatusCode() int {
	return e.Code
}

// Marshal marshals the HTTPError to JSON ignoring possible error.
func (e *HTTPError) Marshal() []byte {
	data, _ := json.Marshal(e)
	return data
}


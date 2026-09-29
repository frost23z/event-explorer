package models

import (
	"errors"
	"net/http"
)

const MsgInternal = "Something went wrong. Please try again."

type APIError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e APIError) Error() string {
	return e.Message
}

func (e APIError) Unwrap() error {
	return e.Err
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func AsAPIError(err error) APIError {
	if apiErr, ok := errors.AsType[APIError](err); ok {
		return apiErr
	}

	return APIError{
		StatusCode: http.StatusInternalServerError,
		Message:    MsgInternal,
		Err:        err,
	}
}

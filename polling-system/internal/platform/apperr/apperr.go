package apperr

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code    string `json:"error"`
	Message string `json:"message"`
	Err     error  `json:"-"`
	status  int
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Code != "" {
		return e.Code
	}
	return e.Err.Error()
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *AppError) StatusCode() int {
	if e == nil || e.status == 0 {
		return http.StatusInternalServerError
	}
	return e.status
}

func BadRequest(code, msg string, err error) *AppError {
	return newAppError(code, msg, err, http.StatusBadRequest)
}

func NotFound(code, msg string, err error) *AppError {
	return newAppError(code, msg, err, http.StatusNotFound)
}

func Conflict(code, msg string, err error) *AppError {
	return newAppError(code, msg, err, http.StatusConflict)
}

func Unauthorized(code, msg string, err error) *AppError {
	return newAppError(code, msg, err, http.StatusUnauthorized)
}

func Forbidden(code, msg string, err error) *AppError {
	return newAppError(code, msg, err, http.StatusForbidden)
}

func Internal(code, msg string, err error) *AppError {
	return newAppError(code, msg, err, http.StatusInternalServerError)
}

func FromError(err error) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return Internal("internal_error", http.StatusText(http.StatusInternalServerError), err)
}

func newAppError(code, msg string, err error, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: msg,
		Err:     err,
		status:  status,
	}
}

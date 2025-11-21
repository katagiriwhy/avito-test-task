package errors

import "strings"

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Code + ": " + e.Message
}

func NewNotFound(code, msg string) *AppError {
	return &AppError{Code: code, Message: msg, Status: 404}
}

func NewConflict(code, msg string) *AppError {
	return &AppError{Code: code, Message: msg, Status: 409}
}

func NewBadRequest(code, msg string) *AppError {
	return &AppError{Code: code, Message: msg, Status: 400}
}

func ValidateString(s string) bool {
	return strings.TrimSpace(s) == ""
}

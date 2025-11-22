package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrBudgetNotFound     = errors.New("budget not found")
	ErrExpenseNotFound    = errors.New("expense not found")
	ErrAlertNotFound      = errors.New("alert not found")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInternalServer     = errors.New("internal server error")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrDatabaseError      = errors.New("database error")
	ErrInvalidObjectID    = errors.New("invalid objectID")
)

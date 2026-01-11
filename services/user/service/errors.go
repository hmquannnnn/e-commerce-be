package service

import "errors"

// Common errors shared across services
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account is locked due to too many failed login attempts")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidPassword    = errors.New("password does not meet requirements")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

package domain

import "errors"

var (
	ErrNotFound           = errors.New("record not found")
	ErrEmailTaken          = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInsufficientStock  = errors.New("insufficient stock")
)


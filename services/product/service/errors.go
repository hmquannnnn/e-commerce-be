package service

import "errors"

var (
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrProductNotFound       = errors.New("product not found")
	ErrInventoryNotFound     = errors.New("inventory not found")
	ErrInsufficientStock     = errors.New("insufficient stock")
	ErrInvalidInput          = errors.New("invalid input")
)

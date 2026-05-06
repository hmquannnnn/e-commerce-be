package service

import "errors"

var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrPaymentNotFound = errors.New("payment not found")
	ErrUnsupported     = errors.New("unsupported provider")
	ErrOrderNotFound   = errors.New("order not found")
	ErrForbidden       = errors.New("forbidden")
	ErrOrderNotPayable = errors.New("order is not in payable state")
	ErrAmountMismatch  = errors.New("amount does not match order total")
)

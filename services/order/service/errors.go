package service

import "errors"

var (
	ErrCartEmpty               = errors.New("cart is empty")
	ErrCartItemNotFound        = errors.New("cart item not found")
	ErrCartChanged             = errors.New("cart changed during checkout")
	ErrOrderNotFound           = errors.New("order not found")
	ErrOrderNotCancellable     = errors.New("order cannot be cancelled in current status")
	ErrInvalidStatusTransition = errors.New("invalid order status transition")
	ErrProductNotFound         = errors.New("product not found")
	ErrInsufficientStock       = errors.New("insufficient stock")
	ErrAmountMismatch          = errors.New("amount does not match order total")
	ErrInvalidInput            = errors.New("invalid input")
	ErrInvalidOrderLines       = errors.New("invalid order lines")
	ErrForbidden               = errors.New("forbidden")
)

package entity

import "errors"

var (
	ErrNotFound              = errors.New("resource not found")
	ErrInvalidInput          = errors.New("invalid input")
	ErrUnauthorized          = errors.New("unauthorized access")
	ErrPlanNotFound          = errors.New("plan not found")
	ErrSubscriptionNotFound  = errors.New("subscription not found")
	ErrInvoiceNotFound       = errors.New("invoice not found")
	ErrPaymentNotFound       = errors.New("payment not found")
	ErrInvalidStatus         = errors.New("invalid status")
	ErrInvalidAmount         = errors.New("invalid amount")
	ErrInvalidInterval       = errors.New("invalid interval")
	ErrAlreadyExists         = errors.New("resource already exists")
	ErrCannotCancel          = errors.New("cannot cancel subscription")
	ErrCannotRefund          = errors.New("cannot refund payment")
	ErrPaymentFailed         = errors.New("payment failed")
)

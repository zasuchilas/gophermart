package common

import "github.com/Rhymond/go-money"

const (
	Currency = money.RUB

	OrderStatusNew        = "NEW"        // gophermart service
	OrderStatusRegistered = "REGISTERED" // accrual service
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

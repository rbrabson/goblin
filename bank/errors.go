package bank

import "errors"

var (
	ErrInsufficientFunds = errors.New("insufficient funds in account for withdrawal")
)

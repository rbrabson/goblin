package bank

import "errors"

var (
	ErrInsufficentFunds = errors.New("insufficient funds in account for withdrawl")
)

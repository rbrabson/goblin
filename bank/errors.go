package bank

import "errors"

var (
	ErrInsufficentFunds    = errors.New("insufficient funds in account for withdrawl")
	ErrUnableToSaveAccount = errors.New("unable to save bank account to the database")
	ErrUnableToSaveBank    = errors.New("unable to save bank to the database")
)

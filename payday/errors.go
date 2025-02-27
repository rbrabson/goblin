package payday

import "errors"

var (
	ErrUnableToSavePayday  = errors.New("unable to save payday to the database")
	ErrUnableToSaveAccount = errors.New("unable to read payday account to the database")
)

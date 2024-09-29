package heist

import "errors"

var (
	ErrConfigNotFound = errors.New("configuration file not found")
	ErrNotAllowed     = errors.New("user is not allowed to perform command")
	ErrNoHeist        = errors.New("heist not found")
)

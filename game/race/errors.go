package race

import "errors"

var (
	ErrConfigNotFound = errors.New("configuration file not found")
	ErrMemberNotFound = errors.New("member not found")
)

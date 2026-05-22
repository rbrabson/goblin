package discord

import "errors"

var (
	ErrAlreadyAdmin        = errors.New("you are already an admin")
	ErrAlreadyOwner        = errors.New("you are already an owner")
	ErrNotAdmin            = errors.New("you are not an admin")
	ErrNotOwner            = errors.New("you are not an owner")
	ErrNoAdmins            = errors.New("there are no admins for this server")
	ErrUnrecognizedCommand = errors.New("unrecognized command")
	ErrNoPermissions       = errors.New("you do not have permission to use this command")
)

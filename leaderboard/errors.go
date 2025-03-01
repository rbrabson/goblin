package leaderboard

import "errors"

var (
	ErrUnableToSaveLeaderboard = errors.New("unable to save leaderboard to the database")
)

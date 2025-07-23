package stats

import "time"

// today returns the current date with the time set to midnight.
func today() time.Time {
	now := time.Now()
	year, month, day := now.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, now.Location())
}

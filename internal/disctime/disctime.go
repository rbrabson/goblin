package disctime

import "time"

// currMon th returns the month and year at for the start of the month
func CurrentMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	month := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	return month
}

// PreviousMonth returns the previous year and month
func PreviousMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	month := time.Date(y, m-1, 1, 0, 0, 0, 0, time.UTC)
	return month
}

// NextMonth returns the next year and month
func NextMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	month := time.Date(y, m+1, 1, 0, 0, 0, 0, time.UTC)
	return month
}

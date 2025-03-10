package disctime

import (
	"strconv"
	"time"
)

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

// ParseDuration parses a duration string.
func ParseDuration(t string) (time.Duration, error) {
	// No string, so return a duration of 0
	runes := []rune(t)
	if len(runes) == 0 {
		return 0, nil
	}

	// Convert years to a duration
	if runes[len(runes)-1] == 'y' || runes[len(runes)-1] == 'Y' {
		years, err := strconv.Atoi(t[:len(t)-1])
		if err != nil {
			return 0, err
		}
		y, m, d := time.Now().Date()
		futureDate := time.Date(y+years, m, d, 0, 0, 0, 0, time.UTC)
		duration := time.Until(futureDate)
		return duration, nil
	}

	// Convert months to a duration
	if runes[len(runes)-1] == 'm' || runes[len(runes)-1] == 'M' {
		months, err := strconv.Atoi(t[:len(t)-1])
		if err != nil {
			return 0, err
		}
		y, m, d := time.Now().Date()
		futureDate := time.Date(y, m+time.Month(months), d, 0, 0, 0, 0, time.UTC)
		duration := time.Until(futureDate)
		return duration, nil
	}

	// Convert days to a duration
	if runes[len(runes)-1] == 'd' || runes[len(runes)-1] == 'D' {
		days, _ := strconv.Atoi(t[:len(t)-1])
		duration := time.Duration(days) * 24 * time.Hour
		return duration, nil
	}

	// Parse as a normal time duration
	return time.ParseDuration(t)
}

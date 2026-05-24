package disctime

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CurrentMonth returns the month and year at for the start of the month
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

// RoundToNextDay rounds the time up to the next whole day. The time is returned
// in UTC.
func RoundToNextDay(t time.Time) time.Time {
	// Round to the next day
	utc := t.UTC()
	year, month, day := utc.Date()
	hour, minute, _ := utc.Clock()
	if hour != 0 || minute != 0 {
		day++
	}

	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// ParseDuration parses a duration string.
func ParseDuration(t string) (time.Duration, error) {
	// No string, so return a duration of 0
	runes := []rune(t)
	if len(runes) == 0 {
		return 0, nil
	}

	var year, month, day int
	pieces := strings.Split(t, " ")
	for _, piece := range pieces {
		runes := []rune(piece)
		if piece == "" {
			continue
		}
		numToAdd, err := strconv.Atoi(string(runes[:len(runes)-1]))
		if err != nil {
			return 0, fmt.Errorf("invalid duration: %s", t)
		}
		switch runes[len(runes)-1] {
		case 'y', 'Y':
			year += numToAdd
		case 'm', 'M':
			month += numToAdd
		case 'd', 'D':
			day += numToAdd
		default:
			return 0, fmt.Errorf("invalid duration: %s", t)
		}
	}

	now := time.Now()
	future := now.AddDate(year, month, day)
	return future.Sub(now), nil
}

// FormatDuration returns duration formatted for inclusion in Discord messages.
func FormatDuration(duration time.Duration) string {
	now := time.Now()
	currentYear, currentMonth, currentDay := now.Date()
	futureYear, futureMonth, futureDay := now.Add(duration).Date()
	elapsedYear := futureYear - currentYear
	elapsedMonth := futureMonth - currentMonth
	elapsedDay := futureDay - currentDay

	sb := strings.Builder{}
	appendDurationPart(&sb, int64(elapsedYear), "year")
	appendDurationPart(&sb, int64(elapsedMonth), "month")
	appendDurationPart(&sb, int64(elapsedDay), "day")
	if sb.Len() > 0 {
		return sb.String()
	}

	remaining := duration.Round(time.Second)
	months := remaining / (time.Hour * 24 * 30)
	remaining -= months * (time.Hour * 24 * 30)
	days := remaining / (time.Hour * 24)
	remaining -= days * (time.Hour * 24)
	hours := remaining / time.Hour
	remaining -= hours * time.Hour
	minutes := remaining / time.Minute
	remaining -= minutes * time.Minute
	seconds := remaining / time.Second

	if months >= 1 {
		return roundedDurationPart(int64(months), int64(days), 15, "month")
	}
	if days >= 1 {
		return roundedDurationPart(int64(days), int64(hours), 12, "day")
	}
	if hours >= 1 {
		return roundedDurationPart(int64(hours), int64(minutes), 30, "hour")
	}
	if minutes >= 1 {
		return roundedDurationPart(int64(minutes), int64(seconds), 30, "minute")
	}
	return durationPart(int64(seconds), "second")
}

// appendDurationPart appends a duration part to the string builder if the value is greater than 0. If the string builder
// already has content, a comma and space are added before the new part.
func appendDurationPart(sb *strings.Builder, value int64, unit string) {
	if value <= 0 {
		return
	}
	if sb.Len() > 0 {
		sb.WriteString(", ")
	}
	sb.WriteString(durationPart(value, unit))
}

// roundedDurationPart rounds the duration part up if the remainder exceeds the roundUpAfter threshold and returns
// the formatted duration part.
func roundedDurationPart(value, remainder, roundUpAfter int64, unit string) string {
	if remainder > roundUpAfter {
		value++
	}
	return durationPart(value, unit)
}

// durationPart returns a formatted duration part. If the value is 1, the unit is singular; otherwise, it is plural.
func durationPart(value int64, unit string) string {
	if value == 1 {
		return fmt.Sprintf("1 %s", unit)
	}
	return fmt.Sprintf("%d %ss", value, unit)
}

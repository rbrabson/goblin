package disctime

import (
	"fmt"
	"strconv"
	"strings"
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

// FormatDuration returns duration formatted for inclusion in Discord messages.
func FormatDuration(duration time.Duration) string {
	currentYear, currentMonth, currentDay := time.Now().Date()
	futureYear, futureMonth, futureDay := time.Now().Add(duration).Date()
	elapsedYear := futureYear - currentYear
	elapsedMonth := futureMonth - currentMonth
	elapsedDay := futureDay - currentDay

	sb := strings.Builder{}
	if elapsedYear > 0 {
		if elapsedYear == 1 {
			sb.WriteString("1 year")
		} else {
			sb.WriteString(fmt.Sprintf("%d years", elapsedYear))
		}
	}
	if elapsedMonth > 0 {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		if elapsedMonth == 1 {
			sb.WriteString("1 month")
		} else {
			sb.WriteString(fmt.Sprintf("%d months", elapsedMonth))
		}
	}
	if elapsedDay > 0 {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		if elapsedDay == 1 {
			sb.WriteString("1 day")
		} else {
			sb.WriteString(fmt.Sprintf("%d days", elapsedDay))
		}
	}
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

	if months == 1 {
		if days <= 15 { // If less than half a month, round down
			return "1 month"
		} // If more than half a month, round up
		return "2 months"
	}
	if months >= 1 {
		if days > 15 {
			months++
		}
		if months == 1 {
			return "1 month"
		}
		return fmt.Sprintf("%d months", months)
	}

	if days == 1 {
		if hours <= 12 { // If less than half a day, round down
			return "1 day"
		} // If more than half a day, round up
		return "2 days"
	}
	if days >= 1 {
		if hours > 12 {
			days++
		}
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}

	if hours == 1 {
		if minutes <= 30 {
			return "1 hour"
		}
		return "2 hours"
	}
	if hours >= 1 {
		if minutes > 30 {
			hours++
		}
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	if minutes >= 1 {
		if seconds > 30 {
			minutes++
		}
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}
	if seconds <= 1 {
		return "1 second"
	}
	return fmt.Sprintf("%d seconds", seconds)
}

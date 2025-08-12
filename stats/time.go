package stats

import (
	"fmt"
	"strings"
	"time"
)

const (
	OneDay       = "one_day"
	OneWeek      = "one_week"
	OneMonth     = "one_month"
	ThreeMonths  = "three_months"
	SixMonths    = "six_months"
	NineMonths   = "nine_months"
	TwelveMonths = "twelve_months"
)

const (
	OneDayAgo       = "one_day_ago"
	LastWeek        = "last_week"
	LastMonth       = "last_month"
	ThreeMonthsAgo  = "three_months_ago"
	SixMonthsAgo    = "six_months_ago"
	NineMonthsAgo   = "nine_months_ago"
	TwelveMonthsAgo = "twelve_months_ago"
)

// today returns the current date with the time set to midnight.
func today() time.Time {
	now := time.Now().UTC()
	year, month, day := now.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, now.Location()).UTC()
}

// getDuration calculates the duration from the start date to today based on the specified period.
func getDuration(guildID string, game string, period string, firstGameDate time.Time) time.Duration {
	today := today().UTC()

	var startDate time.Time
	switch period {
	case OneDay:
		startDate = today.AddDate(0, 0, -1).UTC()
	case OneWeek:
		startDate = today.AddDate(0, 0, -7).UTC()
	case OneMonth:
		startDate = today.AddDate(0, -1, 0).UTC()
	case ThreeMonths:
		startDate = today.AddDate(0, -3, 0).UTC()
	case SixMonths:
		startDate = today.AddDate(0, -6, 0).UTC()
	case NineMonths:
		startDate = today.AddDate(0, -9, 0).UTC()
	case TwelveMonths:
		startDate = today.AddDate(-1, 0, 0).UTC()
	default:
		startDate = firstGameDate
	}

	if firstGameDate.After(startDate) {
		startDate = firstGameDate
	}

	return today.Sub(startDate)
}

// getTime calculates the time period based on the specified period and first game date.
func getTime(period string, firstGameDate time.Time) time.Time {
	today := today().UTC()

	var timePeriod time.Time
	switch period {
	case OneDayAgo:
		timePeriod = today.AddDate(0, 0, -1).UTC()
	case LastWeek:
		timePeriod = today.AddDate(0, 0, -7).UTC()
	case LastMonth:
		timePeriod = today.AddDate(0, -1, 0).UTC()
	case ThreeMonthsAgo:
		timePeriod = today.AddDate(0, -3, 0).UTC()
	case SixMonthsAgo:
		timePeriod = today.AddDate(0, -6, 0).UTC()
	case NineMonthsAgo:
		timePeriod = today.AddDate(0, -9, 0).UTC()
	case TwelveMonthsAgo:
		timePeriod = today.AddDate(-1, 0, 0).UTC()
	default:
		timePeriod = firstGameDate
	}

	if firstGameDate.After(timePeriod) {
		timePeriod = firstGameDate
	}

	if firstGameDate.Equal(timePeriod) {
		timePeriod = timePeriod.AddDate(0, 0, -1).UTC()
	}

	return timePeriod
}

// fmtDuration formats a duration into a string.
func fmtDuration(d time.Duration) string {
	today := today()
	previousDate := today.Add(-d)
	currentYear, currentMonth, currentDay := today.Date()
	previousYear, previousMonth, previousDay := previousDate.Date()
	years := currentYear - previousYear
	months := currentMonth - previousMonth
	days := currentDay - previousDay

	dateParts := make([]string, 0, 3)
	if years > 0 {
		if years == 1 {
			dateParts = append(dateParts, "1 Year")
		} else {
			dateParts = append(dateParts, fmt.Sprintf("%d Years", years))
		}
	}
	if months > 0 {
		if months == 1 {
			dateParts = append(dateParts, "1 Month")
		} else {
			dateParts = append(dateParts, fmt.Sprintf("%d Months", months))
		}
	}
	if days > 0 {
		if days == 1 {
			dateParts = append(dateParts, "1 Day")
		} else {
			dateParts = append(dateParts, fmt.Sprintf("%d Days", days))
		}
	}

	if len(dateParts) == 0 {
		return "Today"
	}

	return strings.Join(dateParts, " ")
}

// timeToString converts a time period string to a human-readable format.
func timeToString(timeString string) string {
	switch timeString {
	case OneDay:
		return "1 Day"
	case OneWeek:
		return "1 Week"
	case OneMonth:
		return "1 Month"
	case ThreeMonths:
		return "3 Months"
	case SixMonths:
		return "6 Months"
	case NineMonths:
		return "9 Months"
	case TwelveMonths:
		return "12 Months"
	case OneDayAgo:
		return "Yesterday"
	case LastWeek:
		return "Last Week"
	case LastMonth:
		return "Last Month"
	case ThreeMonthsAgo:
		return "3 Months ago"
	case SixMonthsAgo:
		return "6 Months ago"
	case NineMonthsAgo:
		return "9 Months ago"
	case TwelveMonthsAgo:
		return "12 Months ago"
	default:
		return ""
	}
}

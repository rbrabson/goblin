package stats

import "time"

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
	return time.Date(year, month, day, 0, 0, 0, 0, now.Location())
}

func getDuration(period string) time.Duration {
	today := today().UTC()
	switch period {
	case OneDay:
		oneDay := today.AddDate(0, 0, -1)
		return today.Sub(oneDay)
	case OneWeek:
		oneWeek := today.AddDate(0, 0, -7)
		return today.Sub(oneWeek)
	case OneMonth:
		oneMonth := today.AddDate(0, -1, 0)
		return today.Sub(oneMonth)
	case ThreeMonths:
		threeMonths := today.AddDate(0, -3, 0)
		return today.Sub(threeMonths)
	case SixMonths:
		sixMonths := today.AddDate(0, -6, 0)
		return today.Sub(sixMonths)
	case NineMonths:
		nineMonths := today.AddDate(0, -9, 0)
		return today.Sub(nineMonths)
	case TwelveMonths:
		twelveMonths := today.AddDate(-1, 0, 0)
		return today.Sub(twelveMonths)
	default:
		return today.Sub(time.Time{})
	}
}

func getTime(period string) time.Time {
	switch period {
	case LastWeek:
		return today().AddDate(0, 0, -7)
	case LastMonth:
		return today().AddDate(0, -1, 0)
	case ThreeMonthsAgo:
		return today().AddDate(0, -3, 0)
	case SixMonthsAgo:
		return today().AddDate(0, -6, 0)
	case NineMonthsAgo:
		return today().AddDate(0, -9, 0)
	case TwelveMonthsAgo:
		return today().AddDate(-1, 0, 0)
	default:
		return time.Time{}
	}
}

func timeToString(timeString string) string {
	switch timeString {
	case OneDay:
		return "1 Eay"
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

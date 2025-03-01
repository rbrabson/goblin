package format

import (
	"fmt"
	"time"
)

// Duration returns duration formatted for inclusion in Discord messages.
func Duration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h == 1 {
		if m <= 30 {
			return "1 hour"
		}
		return "2 hours"
	}
	if h >= 1 {
		if m > 30 {
			h++
		}
		if h == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", h)
	}
	if m >= 1 {
		if s > 30 {
			m++
		}
		if m == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", m)
	}
	if s <= 1 {
		return "1 second"
	}
	return fmt.Sprintf("%d seconds", s)
}

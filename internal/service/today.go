package service

import (
	"time"

	"todoapp/internal/model"
)

// dayOf returns the calendar day of now in the profile's time zone, as a UTC midnight time.
func dayOf(now time.Time, p model.Profile) time.Time {
	y, m, d := now.In(p.Location()).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// dateString formats a day as YYYY-MM-DD.
func dateString(day time.Time) string { return day.Format(model.DateLayout) }

// TodayString returns today's date (YYYY-MM-DD) in the profile's time zone.
func TodayString(now time.Time, p model.Profile) string { return dateString(dayOf(now, p)) }

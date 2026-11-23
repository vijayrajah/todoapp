package service

import (
	"fmt"
	"time"

	"todoapp/internal/model"
)

// NextDue returns the due date of the next occurrence of a recurring task.
// The base is the current due date, or today when the rule counts from completion.
func NextDue(r model.Recurrence, dueDate string, today string) (string, error) {
	base := dueDate
	if r.FromCompletion {
		base = today
	}
	day, err := time.Parse(model.DateLayout, base)
	if err != nil {
		return "", fmt.Errorf("parse date %q: %w", base, err)
	}
	var next time.Time
	switch r.Unit {
	case model.UnitDaily:
		next = day.AddDate(0, 0, r.Interval)
	case model.UnitWeekly:
		next = day.AddDate(0, 0, 7*r.Interval)
	case model.UnitMonthly:
		next = addMonths(day, r.Interval)
	case model.UnitYearly:
		next = addMonths(day, 12*r.Interval)
	default:
		return "", fmt.Errorf("unknown recurrence unit %q", r.Unit)
	}
	return next.Format(model.DateLayout), nil
}

// addMonths adds n months, clamping to the last day of the target month (Jan 31 + 1 month = Feb 28/29).
func addMonths(day time.Time, n int) time.Time {
	y, m, d := day.Date()
	first := time.Date(y, m+time.Month(n), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1).Day()
	if d > last {
		d = last
	}
	return time.Date(first.Year(), first.Month(), d, 0, 0, 0, 0, time.UTC)
}

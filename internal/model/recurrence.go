package model

import (
	"fmt"
	"strconv"
	"strings"
)

// Recurrence unit names.
const (
	UnitDaily   = "daily"
	UnitWeekly  = "weekly"
	UnitMonthly = "monthly"
	UnitYearly  = "yearly"
)

// Recurrence describes how a task repeats. Stored as "unit:interval[:after]".
type Recurrence struct {
	Unit           string
	Interval       int
	FromCompletion bool
}

// ParseRecurrence parses the stored string form.
func ParseRecurrence(s string) (Recurrence, error) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return Recurrence{}, ValidationError("invalid recurrence")
	}
	switch parts[0] {
	case UnitDaily, UnitWeekly, UnitMonthly, UnitYearly:
	default:
		return Recurrence{}, ValidationError("invalid recurrence unit")
	}
	n, err := strconv.Atoi(parts[1])
	if err != nil || n < 1 || n > 365 {
		return Recurrence{}, ValidationError("invalid recurrence interval")
	}
	r := Recurrence{Unit: parts[0], Interval: n}
	if len(parts) == 3 {
		if parts[2] != "after" {
			return Recurrence{}, ValidationError("invalid recurrence mode")
		}
		r.FromCompletion = true
	}
	return r, nil
}

// String returns the stored form.
func (r Recurrence) String() string {
	s := fmt.Sprintf("%s:%d", r.Unit, r.Interval)
	if r.FromCompletion {
		s += ":after"
	}
	return s
}

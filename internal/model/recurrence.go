package model

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Recurrence unit names.
const (
	UnitDaily   = "daily"
	UnitWeekly  = "weekly"
	UnitMonthly = "monthly"
	UnitYearly  = "yearly"
)

var dayCodes = map[string]time.Weekday{
	"mon": time.Monday, "tue": time.Tuesday, "wed": time.Wednesday, "thu": time.Thursday,
	"fri": time.Friday, "sat": time.Saturday, "sun": time.Sunday,
}

// WeekdayIndex is the position in a week that starts on Monday (Mon = 0 ... Sun = 6).
func WeekdayIndex(d time.Weekday) int { return (int(d) + 6) % 7 }

// Recurrence describes how a task repeats. Stored as "unit:interval[:days][:after]",
// where days (weekly only) is a comma list such as "mon,wed,fri".
type Recurrence struct {
	Unit           string
	Interval       int
	Days           []time.Weekday // weekly only; sorted Monday first; empty = same weekday as the due date
	FromCompletion bool
}

// ParseRecurrence parses the stored string form.
func ParseRecurrence(s string) (Recurrence, error) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 4 {
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
	rest := parts[2:]
	if len(rest) > 0 && rest[len(rest)-1] == "after" {
		r.FromCompletion = true
		rest = rest[:len(rest)-1]
	}
	switch len(rest) {
	case 0:
	case 1:
		if r.Unit != UnitWeekly {
			return Recurrence{}, ValidationError("weekdays only apply to weekly recurrence")
		}
		if r.Days, err = parseDays(rest[0]); err != nil {
			return Recurrence{}, err
		}
	default:
		return Recurrence{}, ValidationError("invalid recurrence mode")
	}
	return r, nil
}

func parseDays(s string) ([]time.Weekday, error) {
	seen := map[time.Weekday]bool{}
	for _, code := range strings.Split(s, ",") {
		d, ok := dayCodes[code]
		if !ok {
			return nil, ValidationError("invalid recurrence weekday")
		}
		seen[d] = true
	}
	return SortedDays(seen), nil
}

// SortedDays returns the set as a slice ordered Monday first.
func SortedDays(set map[time.Weekday]bool) []time.Weekday {
	days := make([]time.Weekday, 0, len(set))
	for d := range set {
		days = append(days, d)
	}
	sort.Slice(days, func(i, j int) bool { return WeekdayIndex(days[i]) < WeekdayIndex(days[j]) })
	return days
}

// DayCode returns the stored short name of a weekday ("mon").
func DayCode(d time.Weekday) string { return strings.ToLower(d.String()[:3]) }

// String returns the stored form.
func (r Recurrence) String() string {
	s := fmt.Sprintf("%s:%d", r.Unit, r.Interval)
	if len(r.Days) > 0 {
		codes := make([]string, len(r.Days))
		for i, d := range r.Days {
			codes[i] = DayCode(d)
		}
		s += ":" + strings.Join(codes, ",")
	}
	if r.FromCompletion {
		s += ":after"
	}
	return s
}

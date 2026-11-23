package model

import (
	"errors"
	"time"
)

// ErrNotFound is returned when an entity does not exist.
var ErrNotFound = errors.New("not found")

// ValidationError is a user-correctable input problem.
type ValidationError string

func (e ValidationError) Error() string { return string(e) }

// Time formats used across the app. Timestamps are UTC and fixed-width so they sort as text.
const (
	TimeLayout = "2006-01-02T15:04:05.000Z"
	DateLayout = "2006-01-02"
)

// FormatTime renders t as a UTC timestamp string.
func FormatTime(t time.Time) string { return t.UTC().Format(TimeLayout) }

// ParseTimestamp accepts any RFC3339 time and returns it in the canonical layout.
func ParseTimestamp(s string) (string, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return "", ValidationError("invalid date/time, expected RFC3339")
	}
	return FormatTime(t), nil
}

// ValidDate reports whether s is a YYYY-MM-DD date.
func ValidDate(s string) bool {
	_, err := time.Parse(DateLayout, s)
	return err == nil
}

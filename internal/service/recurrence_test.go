package service

import (
	"testing"

	"todoapp/internal/model"
)

func TestNextDue(t *testing.T) {
	tests := []struct {
		name, rule, due, today, want string
	}{
		{"daily", "daily:1", "2026-03-10", "2026-03-10", "2026-03-11"},
		{"every 2 weeks", "weekly:2", "2026-03-10", "2026-03-10", "2026-03-24"},
		{"month end clamps", "monthly:1", "2026-01-31", "2026-01-31", "2026-02-28"},
		{"leap year clamps", "monthly:1", "2028-01-31", "2028-01-31", "2028-02-29"},
		{"yearly leap day", "yearly:1", "2028-02-29", "2028-02-29", "2029-02-28"},
		{"after completion uses today", "daily:3:after", "2026-03-01", "2026-03-10", "2026-03-13"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rule, err := model.ParseRecurrence(tc.rule)
			if err != nil {
				t.Fatal(err)
			}
			got, err := NextDue(rule, tc.due, tc.today)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}

func TestParseRecurrenceRejectsBadInput(t *testing.T) {
	for _, s := range []string{"", "daily", "hourly:1", "daily:0", "daily:x", "daily:1:later"} {
		if _, err := model.ParseRecurrence(s); err == nil {
			t.Errorf("%q: expected error", s)
		}
	}
}

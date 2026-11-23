package service

import (
	"reflect"
	"testing"
	"time"
)

func TestParseQuickAdd(t *testing.T) {
	today := time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC) // a Wednesday
	tests := []struct {
		name, in string
		want     QuickAdd
	}{
		{"plain", "Buy milk", QuickAdd{Title: "Buy milk"}},
		{"tomorrow priority tag", "Pay rent tomorrow !high #bills",
			QuickAdd{Title: "Pay rent", DueDate: "2026-03-12", Priority: 3, Tags: []string{"bills"}}},
		{"project and ISO date", "Report @Work 2026-04-01",
			QuickAdd{Title: "Report", DueDate: "2026-04-01", Project: "Work"}},
		{"in N days", "Call mom in 3 days", QuickAdd{Title: "Call mom", DueDate: "2026-03-14"}},
		{"recurrence", "Water plants every 2 weeks today",
			QuickAdd{Title: "Water plants", DueDate: "2026-03-11", Recurrence: "weekly:2"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseQuickAdd(tc.in, today)
			if tc.want.Tags == nil {
				got.Tags = nil
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

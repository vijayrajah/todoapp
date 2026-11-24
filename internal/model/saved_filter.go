package model

import "strings"

// Saved filter due-date presets.
const (
	DueAny          = ""
	DueNone         = "none"
	DueOverdue      = "overdue"
	DueToday        = "today"
	DueTodayOverdue = "today_overdue"
	DueNext7        = "next7"
	DueThisWeek     = "this_week"
	DueThisMonth    = "this_month"
)

// Saved filter status values.
const (
	FilterOpen = "todo"
	FilterDone = "done"
	FilterAny  = "any"
)

// SavedFilter is a named, reusable task query, e.g. "high priority in project X due this week".
type SavedFilter struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	ProjectID   int64  `json:"project_id"` // 0 = any
	Tag         string `json:"tag"`
	MinPriority int    `json:"min_priority"`
	Due         string `json:"due"`
	Status      string `json:"status"`
	Position    int64  `json:"position"`
	CreatedAt   string `json:"created_at"`
}

// Clean trims and validates user-editable fields.
func (f *SavedFilter) Clean() error {
	f.Name = strings.TrimSpace(f.Name)
	if f.Name == "" {
		return ValidationError("name is required")
	}
	if len(f.Name) > 100 {
		return ValidationError("name is too long")
	}
	if tags := NormalizeTags([]string{f.Tag}); len(tags) > 0 {
		f.Tag = tags[0]
	} else {
		f.Tag = ""
	}
	if f.MinPriority < PriorityNone || f.MinPriority > PriorityHigh {
		return ValidationError("priority must be between 0 and 3")
	}
	if f.ProjectID < 0 {
		return ValidationError("invalid project")
	}
	switch f.Due {
	case DueAny, DueNone, DueOverdue, DueToday, DueTodayOverdue, DueNext7, DueThisWeek, DueThisMonth:
	default:
		return ValidationError("unknown due date option")
	}
	switch f.Status {
	case "":
		f.Status = FilterOpen
	case FilterOpen, FilterDone, FilterAny:
	default:
		return ValidationError("unknown status")
	}
	return nil
}

package model

import "strings"

// Task status values.
const (
	StatusTodo = "todo"
	StatusDone = "done"
)

// Priority levels.
const (
	PriorityNone   = 0
	PriorityLow    = 1
	PriorityMedium = 2
	PriorityHigh   = 3
)

// Task is a unit of work. Empty strings / zero ids mean "not set".
type Task struct {
	ID              int64    `json:"id"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	DescriptionHTML string   `json:"description_html"`
	Status          string   `json:"status"`
	Priority        int      `json:"priority"`
	DueDate         string   `json:"due_date"`
	ProjectID       int64    `json:"project_id"`
	ProjectName     string   `json:"project_name"`
	ProjectColor    string   `json:"project_color"`
	ParentID        int64    `json:"parent_id"`
	Recurrence      string   `json:"recurrence"`
	Position        int64    `json:"position"`
	Tags            []string `json:"tags"`
	Subtasks        []Task   `json:"subtasks"`
	CompletedAt     string   `json:"completed_at"`
	DeletedAt       string   `json:"deleted_at,omitempty"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

// Clean trims and validates user-editable fields.
func (t *Task) Clean() error {
	t.Title = strings.TrimSpace(t.Title)
	if t.Title == "" {
		return ValidationError("title is required")
	}
	if len(t.Title) > 500 {
		return ValidationError("title is too long")
	}
	if t.Priority < PriorityNone || t.Priority > PriorityHigh {
		return ValidationError("priority must be between 0 and 3")
	}
	if t.DueDate != "" && !ValidDate(t.DueDate) {
		return ValidationError("due date must be YYYY-MM-DD")
	}
	if t.Recurrence != "" {
		r, err := ParseRecurrence(t.Recurrence)
		if err != nil {
			return err
		}
		t.Recurrence = r.String()
		if t.DueDate == "" {
			return ValidationError("recurring tasks need a due date")
		}
	}
	t.Tags = NormalizeTags(t.Tags)
	return nil
}

// TaskFilter selects tasks for listing.
type TaskFilter struct {
	Status          string // "", todo or done
	ProjectID       int64
	Tag             string
	DueFrom         string // inclusive
	DueTo           string // inclusive
	MinPriority     int
	TopLevelOnly    bool
	WithSubtasks    bool   // nest subtasks (implies TopLevelOnly)
	Order           string // manual (default), due, completed
	Limit           int
	IncludeArchived bool
}

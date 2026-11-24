package model

import "strings"

// Limits for checklists.
const (
	MaxChecklistItems = 200
	MaxChecklistText  = 500
)

// ChecklistItem is a lightweight to-do line inside a task.
type ChecklistItem struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

// CleanChecklist trims items and drops empty ones.
func CleanChecklist(items []ChecklistItem) ([]ChecklistItem, error) {
	out := make([]ChecklistItem, 0, len(items))
	for _, it := range items {
		it.Text = strings.TrimSpace(it.Text)
		if it.Text == "" {
			continue
		}
		if len(it.Text) > MaxChecklistText {
			return nil, ValidationError("checklist item is too long")
		}
		out = append(out, it)
	}
	if len(out) > MaxChecklistItems {
		return nil, ValidationError("too many checklist items")
	}
	return out, nil
}

// Limits for templates.
const (
	MaxTemplateSubtasks = 100
)

// Template is a reusable task: using it creates a task with these fields, a checklist and subtasks.
type Template struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    int      `json:"priority"`
	Tags        []string `json:"tags"`
	Checklist   []string `json:"checklist"`
	Subtasks    []string `json:"subtasks"`
	Position    int64    `json:"position"`
	CreatedAt   string   `json:"created_at"`
}

// Clean trims and validates user-editable fields.
func (t *Template) Clean() error {
	t.Name = strings.TrimSpace(t.Name)
	t.Title = strings.TrimSpace(t.Title)
	if t.Name == "" {
		return ValidationError("template name is required")
	}
	if t.Title == "" {
		t.Title = t.Name
	}
	if len(t.Name) > 200 || len(t.Title) > 500 {
		return ValidationError("template name or title is too long")
	}
	if len(t.Description) > 20000 {
		return ValidationError("description is too long")
	}
	if t.Priority < PriorityNone || t.Priority > PriorityHigh {
		return ValidationError("priority must be between 0 and 3")
	}
	t.Tags = NormalizeTags(t.Tags)
	var err error
	if t.Checklist, err = cleanLines(t.Checklist, MaxChecklistItems, MaxChecklistText, "checklist"); err != nil {
		return err
	}
	if t.Subtasks, err = cleanLines(t.Subtasks, MaxTemplateSubtasks, 500, "subtask"); err != nil {
		return err
	}
	return nil
}

func cleanLines(lines []string, maxCount, maxLen int, what string) ([]string, error) {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if len(l) > maxLen {
			return nil, ValidationError(what + " item is too long")
		}
		out = append(out, l)
	}
	if len(out) > maxCount {
		return nil, ValidationError("too many " + what + " items")
	}
	return out, nil
}

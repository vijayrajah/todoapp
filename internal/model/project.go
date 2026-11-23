package model

import (
	"regexp"
	"strings"
)

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// ValidColor reports whether s is a #rrggbb colour.
func ValidColor(s string) bool { return colorPattern.MatchString(s) }

// Project groups tasks.
type Project struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Position    int64  `json:"position"`
	Archived    bool   `json:"archived"`
	TotalTasks  int    `json:"total_tasks"`
	DoneTasks   int    `json:"done_tasks"`
	DeletedAt   string `json:"deleted_at,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// Clean trims and validates user-editable fields.
func (p *Project) Clean() error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return ValidationError("project name is required")
	}
	if len(p.Name) > 200 {
		return ValidationError("project name is too long")
	}
	if p.Color == "" {
		p.Color = "#5b6cf0"
	}
	if !ValidColor(p.Color) {
		return ValidationError("color must be #rrggbb")
	}
	return nil
}

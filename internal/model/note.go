package model

import "strings"

// Note is a free-form markdown document.
type Note struct {
	ID        int64    `json:"id"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	BodyHTML  string   `json:"body_html"`
	Pinned    bool     `json:"pinned"`
	Tags      []string `json:"tags"`
	DeletedAt string   `json:"deleted_at,omitempty"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// Clean trims and validates user-editable fields.
func (n *Note) Clean() error {
	n.Title = strings.TrimSpace(n.Title)
	if n.Title == "" {
		return ValidationError("title is required")
	}
	if len(n.Title) > 500 {
		return ValidationError("title is too long")
	}
	n.Tags = NormalizeTags(n.Tags)
	return nil
}

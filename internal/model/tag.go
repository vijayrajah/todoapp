package model

import (
	"sort"
	"strings"
)

// Tag is a label with usage counts.
type Tag struct {
	Name      string `json:"name"`
	TaskCount int    `json:"task_count"`
	NoteCount int    `json:"note_count"`
}

// NormalizeTags trims, lower-cases, strips a leading '#', and de-duplicates tags.
func NormalizeTags(tags []string) []string {
	seen := make(map[string]bool, len(tags))
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.ToLower(strings.TrimLeft(strings.TrimSpace(t), "#"))
		t = strings.TrimSpace(t)
		if t == "" || len(t) > 50 || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

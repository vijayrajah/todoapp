package model

import "strings"

// Reminder is an in-app alert at a point in time, optionally linked to a task or note.
type Reminder struct {
	ID           int64  `json:"id"`
	Message      string `json:"message"`
	TaskID       int64  `json:"task_id"`
	NoteID       int64  `json:"note_id"`
	RemindAt     string `json:"remind_at"`
	SnoozedUntil string `json:"snoozed_until"`
	DismissedAt  string `json:"dismissed_at"`
	CreatedAt    string `json:"created_at"`
}

// EffectiveTime is when the reminder should next fire.
func (r Reminder) EffectiveTime() string {
	if r.SnoozedUntil != "" {
		return r.SnoozedUntil
	}
	return r.RemindAt
}

// Clean trims and validates user-editable fields; RemindAt is normalised.
func (r *Reminder) Clean() error {
	r.Message = strings.TrimSpace(r.Message)
	if r.Message == "" {
		return ValidationError("message is required")
	}
	if len(r.Message) > 500 {
		return ValidationError("message is too long")
	}
	at, err := ParseTimestamp(r.RemindAt)
	if err != nil {
		return err
	}
	r.RemindAt = at
	return nil
}

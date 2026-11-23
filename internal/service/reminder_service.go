package service

import (
	"context"
	"time"

	"todoapp/internal/model"
	"todoapp/internal/store"
)

// ReminderService implements reminder creation, due checks and snoozing.
type ReminderService struct {
	store *store.Store
	now   func() time.Time
}

// NewReminderService returns a ReminderService.
func NewReminderService(st *store.Store) *ReminderService {
	return &ReminderService{store: st, now: time.Now}
}

// Create validates and stores a reminder.
func (s *ReminderService) Create(ctx context.Context, r model.Reminder) (model.Reminder, error) {
	if err := r.Clean(); err != nil {
		return model.Reminder{}, err
	}
	if r.TaskID != 0 {
		if _, err := s.store.GetTask(ctx, r.TaskID); err != nil {
			return model.Reminder{}, model.ValidationError("unknown task")
		}
	}
	id, err := s.store.CreateReminder(ctx, &r)
	if err != nil {
		return model.Reminder{}, err
	}
	return s.store.GetReminder(ctx, id)
}

// Due returns reminders that should be shown now.
func (s *ReminderService) Due(ctx context.Context) ([]model.Reminder, error) {
	return s.store.DueReminders(ctx, model.FormatTime(s.now()))
}

// Snooze postpones a reminder by minutes, or until an RFC3339 time when until is set.
func (s *ReminderService) Snooze(ctx context.Context, id int64, minutes int, until string) (model.Reminder, error) {
	var at string
	switch {
	case until != "":
		parsed, err := model.ParseTimestamp(until)
		if err != nil {
			return model.Reminder{}, err
		}
		at = parsed
	case minutes > 0 && minutes <= 60*24*365:
		at = model.FormatTime(s.now().Add(time.Duration(minutes) * time.Minute))
	default:
		return model.Reminder{}, model.ValidationError("provide minutes or until")
	}
	if err := s.store.SnoozeReminder(ctx, id, at); err != nil {
		return model.Reminder{}, err
	}
	return s.store.GetReminder(ctx, id)
}

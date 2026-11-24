package service

import (
	"context"
	"time"

	"todoapp/internal/model"
	"todoapp/internal/store"
)

// FilterService manages saved filters and runs them.
type FilterService struct {
	store *store.Store
	now   func() time.Time
}

// NewFilterService returns a FilterService.
func NewFilterService(st *store.Store) *FilterService {
	return &FilterService{store: st, now: time.Now}
}

// Create validates and stores a filter.
func (s *FilterService) Create(ctx context.Context, f model.SavedFilter) (model.SavedFilter, error) {
	if err := s.validate(ctx, &f); err != nil {
		return model.SavedFilter{}, err
	}
	id, err := s.store.CreateFilter(ctx, f)
	if err != nil {
		return model.SavedFilter{}, err
	}
	return s.store.GetFilter(ctx, id)
}

// Update validates and saves a filter.
func (s *FilterService) Update(ctx context.Context, f model.SavedFilter) (model.SavedFilter, error) {
	if err := s.validate(ctx, &f); err != nil {
		return model.SavedFilter{}, err
	}
	if err := s.store.UpdateFilter(ctx, f); err != nil {
		return model.SavedFilter{}, err
	}
	return s.store.GetFilter(ctx, f.ID)
}

func (s *FilterService) validate(ctx context.Context, f *model.SavedFilter) error {
	if err := f.Clean(); err != nil {
		return err
	}
	if f.ProjectID != 0 {
		if _, err := s.store.GetProject(ctx, f.ProjectID); err != nil {
			return model.ValidationError("unknown project")
		}
	}
	return nil
}

// Tasks returns the tasks matching a saved filter, soonest due first.
func (s *FilterService) Tasks(ctx context.Context, id int64) ([]model.Task, error) {
	f, err := s.store.GetFilter(ctx, id)
	if err != nil {
		return nil, err
	}
	profile, err := s.store.GetProfile(ctx)
	if err != nil {
		return nil, err
	}
	tf := ToTaskFilter(f, dayOf(s.now(), profile))
	tasks, err := s.store.ListTasks(ctx, tf)
	if err != nil {
		return nil, err
	}
	renderTasks(tasks)
	return tasks, nil
}

// ToTaskFilter converts a saved filter to a store filter relative to today (a UTC-midnight day).
func ToTaskFilter(f model.SavedFilter, today time.Time) model.TaskFilter {
	tf := model.TaskFilter{ProjectID: f.ProjectID, Tag: f.Tag, MinPriority: f.MinPriority, Order: "due"}
	switch f.Status {
	case model.FilterDone:
		tf.Status = model.StatusDone
	case model.FilterAny:
	default:
		tf.Status = model.StatusTodo
	}
	switch f.Due {
	case model.DueNone:
		tf.NoDue = true
	case model.DueOverdue:
		tf.DueTo = dateString(today.AddDate(0, 0, -1))
	case model.DueToday:
		tf.DueFrom, tf.DueTo = dateString(today), dateString(today)
	case model.DueTodayOverdue:
		tf.DueTo = dateString(today)
	case model.DueNext7:
		tf.DueFrom, tf.DueTo = dateString(today), dateString(today.AddDate(0, 0, 6))
	case model.DueThisWeek:
		monday := today.AddDate(0, 0, -model.WeekdayIndex(today.Weekday()))
		tf.DueFrom, tf.DueTo = dateString(monday), dateString(monday.AddDate(0, 0, 6))
	case model.DueThisMonth:
		first := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
		tf.DueFrom, tf.DueTo = dateString(first), dateString(first.AddDate(0, 1, -1))
	}
	return tf
}

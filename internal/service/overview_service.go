package service

import (
	"context"
	"time"

	"todoapp/internal/model"
	"todoapp/internal/store"
)

// Overview is the content of the home page.
type Overview struct {
	Today        string           `json:"today"`
	Overdue      []model.Task     `json:"overdue"`
	DueToday     []model.Task     `json:"due_today"`
	HighPriority []model.Task     `json:"high_priority"`
	Reminders    []model.Reminder `json:"reminders"`
}

// OverviewService builds the overview page data.
type OverviewService struct {
	store     *store.Store
	reminders *ReminderService
	now       func() time.Time
}

// NewOverviewService returns an OverviewService.
func NewOverviewService(st *store.Store, reminders *ReminderService) *OverviewService {
	return &OverviewService{store: st, reminders: reminders, now: time.Now}
}

// Get returns overdue and due-today tasks, high priority tasks not already listed, and due reminders.
func (s *OverviewService) Get(ctx context.Context) (Overview, error) {
	profile, err := s.store.GetProfile(ctx)
	if err != nil {
		return Overview{}, err
	}
	today := dayOf(s.now(), profile)
	todayStr := dateString(today)

	overdue, err := s.store.ListTasks(ctx, model.TaskFilter{
		Status: model.StatusTodo, DueTo: dateString(today.AddDate(0, 0, -1)), Order: "due"})
	if err != nil {
		return Overview{}, err
	}
	dueToday, err := s.store.ListTasks(ctx, model.TaskFilter{
		Status: model.StatusTodo, DueFrom: todayStr, DueTo: todayStr, Order: "due"})
	if err != nil {
		return Overview{}, err
	}
	high, err := s.store.ListTasks(ctx, model.TaskFilter{
		Status: model.StatusTodo, MinPriority: model.PriorityHigh, Order: "due"})
	if err != nil {
		return Overview{}, err
	}
	reminders, err := s.reminders.Due(ctx)
	if err != nil {
		return Overview{}, err
	}

	high = withoutDueBy(high, todayStr)
	for _, list := range [][]model.Task{overdue, dueToday, high} {
		renderTasks(list)
	}
	return Overview{Today: todayStr, Overdue: overdue, DueToday: dueToday, HighPriority: high, Reminders: reminders}, nil
}

// withoutDueBy drops tasks due on or before day; they already appear in the overdue/today lists.
func withoutDueBy(tasks []model.Task, day string) []model.Task {
	kept := []model.Task{}
	for _, t := range tasks {
		if t.DueDate == "" || t.DueDate > day {
			kept = append(kept, t)
		}
	}
	return kept
}

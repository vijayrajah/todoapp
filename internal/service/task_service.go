// Package service holds business rules on top of the store.
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"todoapp/internal/model"
	"todoapp/internal/store"
)

// TaskService implements task rules: views, subtasks, completion and recurrence.
type TaskService struct {
	store *store.Store
	now   func() time.Time
}

// NewTaskService returns a TaskService.
func NewTaskService(st *store.Store) *TaskService {
	return &TaskService{store: st, now: time.Now}
}

func (s *TaskService) today(ctx context.Context) (time.Time, error) {
	profile, err := s.store.GetProfile(ctx)
	if err != nil {
		return time.Time{}, err
	}
	return dayOf(s.now(), profile), nil
}

// List returns tasks for a view: all, today, upcoming or completed.
func (s *TaskService) List(ctx context.Context, view string, projectID int64, tag string) ([]model.Task, error) {
	today, err := s.today(ctx)
	if err != nil {
		return nil, err
	}
	f := model.TaskFilter{ProjectID: projectID, Tag: tag}
	switch view {
	case "", "all":
		f.Status = model.StatusTodo
		f.WithSubtasks = tag == ""
	case "today":
		f.Status = model.StatusTodo
		f.DueTo = dateString(today)
		f.Order = "due"
	case "upcoming":
		f.Status = model.StatusTodo
		f.DueFrom = dateString(today.AddDate(0, 0, 1))
		f.DueTo = dateString(today.AddDate(0, 0, 7))
		f.Order = "due"
	case "completed":
		f.Status = model.StatusDone
		f.TopLevelOnly = true
		f.Order = "completed"
		f.Limit = 200
	default:
		return nil, model.ValidationError("unknown view")
	}
	tasks, err := s.store.ListTasks(ctx, f)
	if err != nil {
		return nil, err
	}
	renderTasks(tasks)
	return tasks, nil
}

// maxRangeDays bounds calendar queries.
const maxRangeDays = 70

// ListRange returns open and done tasks due between from and to (inclusive), flat, for the calendar.
func (s *TaskService) ListRange(ctx context.Context, from, to string) ([]model.Task, error) {
	start, err1 := time.Parse(model.DateLayout, from)
	end, err2 := time.Parse(model.DateLayout, to)
	if err1 != nil || err2 != nil {
		return nil, model.ValidationError("from and to must be YYYY-MM-DD")
	}
	if end.Before(start) || end.Sub(start) > maxRangeDays*24*time.Hour {
		return nil, model.ValidationError("invalid date range")
	}
	tasks, err := s.store.ListTasks(ctx, model.TaskFilter{DueFrom: from, DueTo: to, Order: "due", Limit: 2000})
	if err != nil {
		return nil, err
	}
	renderTasks(tasks)
	return tasks, nil
}

// Get returns one task with subtasks.
func (s *TaskService) Get(ctx context.Context, id int64) (model.Task, error) {
	t, err := s.store.GetTask(ctx, id)
	if err != nil {
		return model.Task{}, err
	}
	renderTask(&t)
	return t, nil
}

// Create validates and stores a task. Subtasks inherit the parent's project.
func (s *TaskService) Create(ctx context.Context, t model.Task) (model.Task, error) {
	if t.ParentID != 0 {
		parent, err := s.store.GetTask(ctx, t.ParentID)
		if err != nil {
			return model.Task{}, err
		}
		if parent.ParentID != 0 {
			return model.Task{}, model.ValidationError("subtasks cannot have subtasks")
		}
		if t.Recurrence != "" {
			return model.Task{}, model.ValidationError("subtasks cannot repeat")
		}
		t.ProjectID = parent.ProjectID
	}
	if err := t.Clean(); err != nil {
		return model.Task{}, err
	}
	t.Subtasks = nil
	id, err := s.store.CreateTask(ctx, &t)
	if err != nil {
		return model.Task{}, err
	}
	return s.Get(ctx, id)
}

// QuickAdd creates a task from a quick-add line. A non-zero projectID overrides any @project.
func (s *TaskService) QuickAdd(ctx context.Context, text string, projectID int64) (model.Task, error) {
	today, err := s.today(ctx)
	if err != nil {
		return model.Task{}, err
	}
	qa := ParseQuickAdd(text, today)

	t := model.Task{
		Title: qa.Title, DueDate: qa.DueDate, Priority: qa.Priority,
		Tags: qa.Tags, Recurrence: qa.Recurrence, ProjectID: projectID,
	}
	if qa.Project != "" && projectID == 0 {
		project, err := s.store.FindProjectByName(ctx, qa.Project)
		switch {
		case err == nil:
			t.ProjectID = project.ID
		case err == model.ErrNotFound:
			t.Title = strings.TrimSpace(t.Title + " @" + qa.Project)
		default:
			return model.Task{}, err
		}
	}
	if t.Recurrence != "" && t.DueDate == "" {
		t.DueDate = dateString(today)
		if rule, err := model.ParseRecurrence(t.Recurrence); err == nil && len(rule.Days) > 0 {
			t.DueDate = dateString(firstOnOrAfter(today, rule.Days))
		}
	}
	return s.Create(ctx, t)
}

// Update saves edits to a task. Parent and (for subtasks) project cannot be changed here.
func (s *TaskService) Update(ctx context.Context, t model.Task) (model.Task, error) {
	existing, err := s.store.GetTask(ctx, t.ID)
	if err != nil {
		return model.Task{}, err
	}
	t.ParentID = existing.ParentID
	if existing.ParentID != 0 {
		if t.Recurrence != "" {
			return model.Task{}, model.ValidationError("subtasks cannot repeat")
		}
		t.ProjectID = existing.ProjectID
	}
	if err := t.Clean(); err != nil {
		return model.Task{}, err
	}
	if err := s.store.UpdateTask(ctx, &t); err != nil {
		return model.Task{}, err
	}
	return s.Get(ctx, t.ID)
}

// Complete marks a task done. A recurring task also gets its next occurrence, returned as next.
func (s *TaskService) Complete(ctx context.Context, id int64) (task model.Task, next *model.Task, err error) {
	existing, err := s.store.GetTask(ctx, id)
	if err != nil {
		return model.Task{}, nil, err
	}

	var nextTask *model.Task
	if existing.Status == model.StatusTodo && existing.Recurrence != "" {
		nextTask, err = s.nextOccurrence(ctx, existing)
		if err != nil {
			return model.Task{}, nil, err
		}
	}
	nextID, err := s.store.CompleteTask(ctx, id, nextTask)
	if err != nil {
		return model.Task{}, nil, err
	}
	task, err = s.Get(ctx, id)
	if err != nil {
		return model.Task{}, nil, err
	}
	if nextID != 0 {
		n, err := s.Get(ctx, nextID)
		if err != nil {
			return model.Task{}, nil, err
		}
		next = &n
	}
	return task, next, nil
}

// nextOccurrence builds the follow-up task for a recurring task, with fresh open subtasks.
func (s *TaskService) nextOccurrence(ctx context.Context, done model.Task) (*model.Task, error) {
	rule, err := model.ParseRecurrence(done.Recurrence)
	if err != nil {
		return nil, fmt.Errorf("task %d: %w", done.ID, err)
	}
	today, err := s.today(ctx)
	if err != nil {
		return nil, err
	}
	due, err := NextDue(rule, done.DueDate, dateString(today))
	if err != nil {
		return nil, err
	}
	next := &model.Task{
		Title: done.Title, Description: done.Description, Priority: done.Priority,
		DueDate: due, ProjectID: done.ProjectID, Recurrence: done.Recurrence, Tags: done.Tags,
	}
	for _, it := range done.Checklist {
		next.Checklist = append(next.Checklist, model.ChecklistItem{Text: it.Text})
	}
	for _, sub := range done.Subtasks {
		next.Subtasks = append(next.Subtasks, model.Task{
			Title: sub.Title, Description: sub.Description, Priority: sub.Priority, Tags: sub.Tags,
		})
	}
	return next, nil
}

// Reopen marks a task as not done.
func (s *TaskService) Reopen(ctx context.Context, id int64) (model.Task, error) {
	if err := s.store.ReopenTask(ctx, id); err != nil {
		return model.Task{}, err
	}
	return s.Get(ctx, id)
}

// Delete moves a task to the trash.
func (s *TaskService) Delete(ctx context.Context, id int64) error {
	return s.store.DeleteTask(ctx, id)
}

// Reorder applies a manual order.
func (s *TaskService) Reorder(ctx context.Context, ids []int64) error {
	return s.store.ReorderTasks(ctx, ids)
}

func renderTasks(tasks []model.Task) {
	for i := range tasks {
		renderTask(&tasks[i])
	}
}

func renderTask(t *model.Task) {
	t.DescriptionHTML = RenderMarkdown(t.Description)
	renderTasks(t.Subtasks)
}

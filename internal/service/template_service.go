package service

import (
	"context"

	"todoapp/internal/model"
	"todoapp/internal/store"
)

// TemplateService manages task templates and creates tasks from them.
type TemplateService struct {
	store *store.Store
}

// NewTemplateService returns a TemplateService.
func NewTemplateService(st *store.Store) *TemplateService {
	return &TemplateService{store: st}
}

// Create validates and stores a template.
func (s *TemplateService) Create(ctx context.Context, t model.Template) (model.Template, error) {
	if err := t.Clean(); err != nil {
		return model.Template{}, err
	}
	id, err := s.store.CreateTemplate(ctx, t)
	if err != nil {
		return model.Template{}, err
	}
	return s.store.GetTemplate(ctx, id)
}

// Update validates and saves a template.
func (s *TemplateService) Update(ctx context.Context, t model.Template) (model.Template, error) {
	if err := t.Clean(); err != nil {
		return model.Template{}, err
	}
	if err := s.store.UpdateTemplate(ctx, t); err != nil {
		return model.Template{}, err
	}
	return s.store.GetTemplate(ctx, t.ID)
}

// UseInput customises the task created from a template.
type UseInput struct {
	ProjectID int64  `json:"project_id"`
	DueDate   string `json:"due_date"`
}

// Use creates a task (with checklist and subtasks) from a template.
func (s *TemplateService) Use(ctx context.Context, id int64, in UseInput) (model.Task, error) {
	tpl, err := s.store.GetTemplate(ctx, id)
	if err != nil {
		return model.Task{}, err
	}
	if in.ProjectID != 0 {
		if _, err := s.store.GetProject(ctx, in.ProjectID); err != nil {
			return model.Task{}, model.ValidationError("unknown project")
		}
	}
	t := model.Task{
		Title: tpl.Title, Description: tpl.Description, Priority: tpl.Priority,
		Tags: tpl.Tags, DueDate: in.DueDate, ProjectID: in.ProjectID,
	}
	for _, text := range tpl.Checklist {
		t.Checklist = append(t.Checklist, model.ChecklistItem{Text: text})
	}
	if err := t.Clean(); err != nil {
		return model.Task{}, err
	}
	for _, title := range tpl.Subtasks {
		sub := model.Task{Title: title}
		if err := sub.Clean(); err != nil {
			return model.Task{}, err
		}
		t.Subtasks = append(t.Subtasks, sub)
	}
	newID, err := s.store.CreateTask(ctx, &t)
	if err != nil {
		return model.Task{}, err
	}
	created, err := s.store.GetTask(ctx, newID)
	if err != nil {
		return model.Task{}, err
	}
	renderTask(&created)
	return created, nil
}

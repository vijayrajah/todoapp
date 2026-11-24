package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"todoapp/internal/model"
)

// templateData is the JSON stored in task_templates.data.
type templateData struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    int      `json:"priority"`
	Tags        []string `json:"tags"`
	Checklist   []string `json:"checklist"`
	Subtasks    []string `json:"subtasks"`
}

func templateJSON(t model.Template) (string, error) {
	b, err := json.Marshal(templateData{t.Title, t.Description, t.Priority, t.Tags, t.Checklist, t.Subtasks})
	return string(b), err
}

func scanTemplate(row scanner) (model.Template, error) {
	var t model.Template
	var raw string
	if err := row.Scan(&t.ID, &t.Name, &raw, &t.Position, &t.CreatedAt); err != nil {
		return t, err
	}
	var d templateData
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return t, err
	}
	t.Title, t.Description, t.Priority = d.Title, d.Description, d.Priority
	t.Tags, t.Checklist, t.Subtasks = nonNil(d.Tags), nonNil(d.Checklist), nonNil(d.Subtasks)
	return t, nil
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

const templateSelect = "SELECT id, name, data, position, created_at FROM task_templates"

// ListTemplates returns all templates in manual order.
func (s *Store) ListTemplates(ctx context.Context) ([]model.Template, error) {
	rows, err := s.db.QueryContext(ctx, templateSelect+" ORDER BY position, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Template{}
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// GetTemplate returns one template.
func (s *Store) GetTemplate(ctx context.Context, id int64) (model.Template, error) {
	t, err := scanTemplate(s.db.QueryRowContext(ctx, templateSelect+" WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return t, model.ErrNotFound
	}
	return t, err
}

// CreateTemplate inserts a template and returns its id.
func (s *Store) CreateTemplate(ctx context.Context, t model.Template) (int64, error) {
	data, err := templateJSON(t)
	if err != nil {
		return 0, err
	}
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO task_templates (name, data, position, created_at) VALUES (?, ?, (SELECT COALESCE(MAX(position), 0) + 1 FROM task_templates), ?)",
		t.Name, data, s.ts())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateTemplate saves a template.
func (s *Store) UpdateTemplate(ctx context.Context, t model.Template) error {
	data, err := templateJSON(t)
	if err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, "UPDATE task_templates SET name = ?, data = ? WHERE id = ?", t.Name, data, t.ID)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

// DeleteTemplate removes a template.
func (s *Store) DeleteTemplate(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM task_templates WHERE id = ?", id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"todoapp/internal/model"
)

// filterData is the JSON stored in saved_filters.data.
type filterData struct {
	ProjectID   int64  `json:"project_id"`
	Tag         string `json:"tag"`
	MinPriority int    `json:"min_priority"`
	Due         string `json:"due"`
	Status      string `json:"status"`
}

func scanFilter(row scanner) (model.SavedFilter, error) {
	var f model.SavedFilter
	var raw string
	if err := row.Scan(&f.ID, &f.Name, &raw, &f.Position, &f.CreatedAt); err != nil {
		return f, err
	}
	var d filterData
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return f, err
	}
	f.ProjectID, f.Tag, f.MinPriority, f.Due, f.Status = d.ProjectID, d.Tag, d.MinPriority, d.Due, d.Status
	return f, nil
}

func filterJSON(f model.SavedFilter) (string, error) {
	b, err := json.Marshal(filterData{f.ProjectID, f.Tag, f.MinPriority, f.Due, f.Status})
	return string(b), err
}

const filterSelect = "SELECT id, name, data, position, created_at FROM saved_filters"

// ListFilters returns all saved filters in manual order.
func (s *Store) ListFilters(ctx context.Context) ([]model.SavedFilter, error) {
	rows, err := s.db.QueryContext(ctx, filterSelect+" ORDER BY position, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.SavedFilter{}
	for rows.Next() {
		f, err := scanFilter(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// GetFilter returns one saved filter.
func (s *Store) GetFilter(ctx context.Context, id int64) (model.SavedFilter, error) {
	f, err := scanFilter(s.db.QueryRowContext(ctx, filterSelect+" WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return f, model.ErrNotFound
	}
	return f, err
}

// CreateFilter inserts a saved filter and returns its id.
func (s *Store) CreateFilter(ctx context.Context, f model.SavedFilter) (int64, error) {
	data, err := filterJSON(f)
	if err != nil {
		return 0, err
	}
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO saved_filters (name, data, position, created_at) VALUES (?, ?, (SELECT COALESCE(MAX(position), 0) + 1 FROM saved_filters), ?)",
		f.Name, data, s.ts())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateFilter saves a saved filter.
func (s *Store) UpdateFilter(ctx context.Context, f model.SavedFilter) error {
	data, err := filterJSON(f)
	if err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, "UPDATE saved_filters SET name = ?, data = ? WHERE id = ?", f.Name, data, f.ID)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

// DeleteFilter removes a saved filter.
func (s *Store) DeleteFilter(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM saved_filters WHERE id = ?", id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

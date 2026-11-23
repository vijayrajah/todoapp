package store

import (
	"context"
	"database/sql"
	"strings"

	"todoapp/internal/model"
)

const projectSelect = `
	SELECT p.id, p.name, p.description, p.color, p.position, p.archived, COALESCE(p.deleted_at, ''), p.created_at,
	  (SELECT COUNT(*) FROM tasks t WHERE t.project_id = p.id AND t.deleted_at IS NULL),
	  (SELECT COUNT(*) FROM tasks t WHERE t.project_id = p.id AND t.deleted_at IS NULL AND t.status = 'done')
	FROM projects p`

func scanProject(row scanner) (model.Project, error) {
	var p model.Project
	var archived int
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Color, &p.Position, &archived, &p.DeletedAt, &p.CreatedAt,
		&p.TotalTasks, &p.DoneTasks)
	p.Archived = archived == 1
	return p, err
}

func queryProjects(ctx context.Context, q querier, where string, args []any, orderBy string) ([]model.Project, error) {
	rows, err := q.QueryContext(ctx, projectSelect+" WHERE "+where+" ORDER BY "+orderBy, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := []model.Project{}
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// ListProjects returns non-trashed projects, either active or archived.
func (s *Store) ListProjects(ctx context.Context, archived bool) ([]model.Project, error) {
	return queryProjects(ctx, s.db, "p.deleted_at IS NULL AND p.archived = ?", []any{boolInt(archived)}, "p.position, p.id")
}

// GetProject returns one non-trashed project.
func (s *Store) GetProject(ctx context.Context, id int64) (model.Project, error) {
	projects, err := queryProjects(ctx, s.db, "p.id = ? AND p.deleted_at IS NULL", []any{id}, "p.id")
	if err != nil {
		return model.Project{}, err
	}
	if len(projects) == 0 {
		return model.Project{}, model.ErrNotFound
	}
	return projects[0], nil
}

// FindProjectByName looks up an active project by name, ignoring case.
func (s *Store) FindProjectByName(ctx context.Context, name string) (model.Project, error) {
	projects, err := queryProjects(ctx, s.db,
		"lower(p.name) = ? AND p.deleted_at IS NULL AND p.archived = 0", []any{strings.ToLower(name)}, "p.id")
	if err != nil {
		return model.Project{}, err
	}
	if len(projects) == 0 {
		return model.Project{}, model.ErrNotFound
	}
	return projects[0], nil
}

// CreateProject inserts a project at the end of the list and returns its id.
func (s *Store) CreateProject(ctx context.Context, p *model.Project) (int64, error) {
	var id int64
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var position int64
		if err := tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(position), 0) + 1 FROM projects").Scan(&position); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx,
			"INSERT INTO projects (name, description, color, position, created_at) VALUES (?, ?, ?, ?, ?)",
			p.Name, p.Description, p.Color, position, s.ts())
		if err != nil {
			return err
		}
		if id, err = res.LastInsertId(); err != nil {
			return err
		}
		return indexProject(ctx, tx, id)
	})
	return id, err
}

// UpdateProject saves name, description and colour.
func (s *Store) UpdateProject(ctx context.Context, p *model.Project) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx,
			"UPDATE projects SET name = ?, description = ?, color = ? WHERE id = ? AND deleted_at IS NULL",
			p.Name, p.Description, p.Color, p.ID)
		if err != nil {
			return err
		}
		if err := requireAffected(res); err != nil {
			return err
		}
		return indexProject(ctx, tx, p.ID)
	})
}

// SetProjectArchived archives or unarchives a project.
func (s *Store) SetProjectArchived(ctx context.Context, id int64, archived bool) error {
	res, err := s.db.ExecContext(ctx, "UPDATE projects SET archived = ? WHERE id = ? AND deleted_at IS NULL", boolInt(archived), id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

// DeleteProject moves a project and its tasks to the trash with one shared stamp.
func (s *Store) DeleteProject(ctx context.Context, id int64) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		stamp := s.ts()
		res, err := tx.ExecContext(ctx, "UPDATE projects SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL", stamp, id)
		if err != nil {
			return err
		}
		if err := requireAffected(res); err != nil {
			return err
		}
		taskIDs, err := idsWhere(ctx, tx, "SELECT id FROM tasks WHERE project_id = ? AND deleted_at IS NULL", id)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "UPDATE tasks SET deleted_at = ? WHERE project_id = ? AND deleted_at IS NULL", stamp, id); err != nil {
			return err
		}
		for _, tid := range taskIDs {
			if err := removeIndex(ctx, tx, "task", tid); err != nil {
				return err
			}
		}
		return removeIndex(ctx, tx, "project", id)
	})
}

// ReorderProjects applies a manual order to the given project ids.
func (s *Store) ReorderProjects(ctx context.Context, ids []int64) error {
	return s.reorder(ctx, "projects", ids)
}

package store

import (
	"context"
	"database/sql"

	"todoapp/internal/model"
)

// ListTrash returns the roots of trashed groups: a task trashed together with its
// project or parent is restored through that project or parent.
func (s *Store) ListTrash(ctx context.Context) (model.Trash, error) {
	tasks, err := queryTasks(ctx, s.db, `t.deleted_at IS NOT NULL
		AND NOT EXISTS (SELECT 1 FROM tasks pt WHERE pt.id = t.parent_id AND pt.deleted_at = t.deleted_at)
		AND NOT EXISTS (SELECT 1 FROM projects pp WHERE pp.id = t.project_id AND pp.deleted_at = t.deleted_at)`,
		nil, "t.deleted_at DESC, t.id", 0)
	if err != nil {
		return model.Trash{}, err
	}
	notes, err := queryNotes(ctx, s.db, "n.deleted_at IS NOT NULL", nil, "n.deleted_at DESC, n.id")
	if err != nil {
		return model.Trash{}, err
	}
	projects, err := queryProjects(ctx, s.db, "p.deleted_at IS NOT NULL", nil, "p.deleted_at DESC, p.id")
	if err != nil {
		return model.Trash{}, err
	}
	return model.Trash{Tasks: tasks, Notes: notes, Projects: projects}, nil
}

// deletedStamp returns the deleted_at of a trashed row; table is a trusted constant.
func deletedStamp(ctx context.Context, tx *sql.Tx, table string, id int64) (string, error) {
	var stamp sql.NullString
	err := tx.QueryRowContext(ctx, "SELECT deleted_at FROM "+table+" WHERE id = ?", id).Scan(&stamp)
	if err == sql.ErrNoRows || (err == nil && !stamp.Valid) {
		return "", model.ErrNotFound
	}
	return stamp.String, err
}

// RestoreTask restores a task and the subtasks trashed with it.
func (s *Store) RestoreTask(ctx context.Context, id int64) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		stamp, err := deletedStamp(ctx, tx, "tasks", id)
		if err != nil {
			return err
		}
		ids, err := idsWhere(ctx, tx, "SELECT id FROM tasks WHERE id = ? OR (parent_id = ? AND deleted_at = ?)", id, id, stamp)
		if err != nil {
			return err
		}
		marks, args := inClause(ids)
		if _, err := tx.ExecContext(ctx, "UPDATE tasks SET deleted_at = NULL WHERE id IN ("+marks+")", args...); err != nil {
			return err
		}
		// Detach from a project or parent that is still in the trash.
		if _, err := tx.ExecContext(ctx,
			"UPDATE tasks SET project_id = NULL WHERE id IN ("+marks+") AND project_id IN (SELECT id FROM projects WHERE deleted_at IS NOT NULL)", args...); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			"UPDATE tasks SET parent_id = NULL WHERE id = ? AND parent_id IN (SELECT id FROM tasks WHERE deleted_at IS NOT NULL)", id); err != nil {
			return err
		}
		for _, tid := range ids {
			if err := indexTask(ctx, tx, tid); err != nil {
				return err
			}
		}
		return nil
	})
}

// RestoreProject restores a project and the tasks trashed with it.
func (s *Store) RestoreProject(ctx context.Context, id int64) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		stamp, err := deletedStamp(ctx, tx, "projects", id)
		if err != nil {
			return err
		}
		ids, err := idsWhere(ctx, tx, "SELECT id FROM tasks WHERE project_id = ? AND deleted_at = ?", id, stamp)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "UPDATE projects SET deleted_at = NULL WHERE id = ?", id); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "UPDATE tasks SET deleted_at = NULL WHERE project_id = ? AND deleted_at = ?", id, stamp); err != nil {
			return err
		}
		if err := indexProject(ctx, tx, id); err != nil {
			return err
		}
		for _, tid := range ids {
			if err := indexTask(ctx, tx, tid); err != nil {
				return err
			}
		}
		return nil
	})
}

// RestoreNote restores a trashed note.
func (s *Store) RestoreNote(ctx context.Context, id int64) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, "UPDATE notes SET deleted_at = NULL WHERE id = ? AND deleted_at IS NOT NULL", id)
		if err != nil {
			return err
		}
		if err := requireAffected(res); err != nil {
			return err
		}
		return indexNote(ctx, tx, id)
	})
}

// PurgeTrashed permanently deletes one trashed item. kind is task, note or project.
func (s *Store) PurgeTrashed(ctx context.Context, kind string, id int64) error {
	table, ok := map[string]string{"task": "tasks", "note": "notes", "project": "projects"}[kind]
	if !ok {
		return model.ValidationError("unknown kind")
	}
	res, err := s.db.ExecContext(ctx, "DELETE FROM "+table+" WHERE id = ? AND deleted_at IS NOT NULL", id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

// EmptyTrash permanently deletes everything in the trash.
func (s *Store) EmptyTrash(ctx context.Context) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		for _, table := range []string{"projects", "tasks", "notes"} {
			if _, err := tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE deleted_at IS NOT NULL"); err != nil {
				return err
			}
		}
		return nil
	})
}

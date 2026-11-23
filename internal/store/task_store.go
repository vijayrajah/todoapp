package store

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"todoapp/internal/model"
)

const taskSelect = `
	SELECT t.id, t.title, t.description, t.status, t.priority,
	  COALESCE(t.due_date, ''), COALESCE(t.project_id, 0), COALESCE(p.name, ''), COALESCE(p.color, ''),
	  COALESCE(t.parent_id, 0), COALESCE(t.recurrence, ''), t.position,
	  COALESCE(t.completed_at, ''), COALESCE(t.deleted_at, ''), t.created_at, t.updated_at
	FROM tasks t LEFT JOIN projects p ON p.id = t.project_id`

func scanTask(row scanner) (model.Task, error) {
	var t model.Task
	err := row.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.DueDate, &t.ProjectID, &t.ProjectName, &t.ProjectColor,
		&t.ParentID, &t.Recurrence, &t.Position,
		&t.CompletedAt, &t.DeletedAt, &t.CreatedAt, &t.UpdatedAt)
	t.Tags = []string{}
	t.Subtasks = []model.Task{}
	return t, err
}

// queryTasks runs the task select with a WHERE and ORDER BY clause and loads tags.
func queryTasks(ctx context.Context, q querier, where string, args []any, orderBy string, limit int) ([]model.Task, error) {
	query := taskSelect + " WHERE " + where + " ORDER BY " + orderBy
	if limit > 0 {
		query += " LIMIT " + strconv.Itoa(limit)
	}
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	tasks := []model.Task{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		tasks = append(tasks, t)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	ids := make([]int64, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}
	tags, err := loadTags(ctx, q, "task_tags", "task_id", ids)
	if err != nil {
		return nil, err
	}
	for i := range tasks {
		if tg := tags[tasks[i].ID]; tg != nil {
			tasks[i].Tags = tg
		}
	}
	return tasks, nil
}

func orderClause(order string) string {
	switch order {
	case "due":
		return "t.due_date IS NULL, t.due_date, t.priority DESC, t.position, t.id"
	case "completed":
		return "t.completed_at DESC, t.id DESC"
	default:
		return "t.position, t.id"
	}
}

// ListTasks returns non-trashed tasks matching the filter.
func (s *Store) ListTasks(ctx context.Context, f model.TaskFilter) ([]model.Task, error) {
	where := []string{"t.deleted_at IS NULL"}
	var args []any
	if f.ProjectID == 0 && !f.IncludeArchived {
		where = append(where, "(p.id IS NULL OR p.archived = 0)")
	}
	if f.Status != "" {
		where = append(where, "t.status = ?")
		args = append(args, f.Status)
	}
	if f.ProjectID != 0 {
		where = append(where, "t.project_id = ?")
		args = append(args, f.ProjectID)
	}
	if f.Tag != "" {
		where = append(where, `EXISTS (SELECT 1 FROM task_tags tt JOIN tags g ON g.id = tt.tag_id WHERE tt.task_id = t.id AND g.name = ?)`)
		args = append(args, f.Tag)
	}
	if f.DueFrom != "" {
		where = append(where, "t.due_date >= ?")
		args = append(args, f.DueFrom)
	}
	if f.DueTo != "" {
		where = append(where, "t.due_date <= ?")
		args = append(args, f.DueTo)
	}
	if f.MinPriority > 0 {
		where = append(where, "t.priority >= ?")
		args = append(args, f.MinPriority)
	}
	if f.TopLevelOnly || f.WithSubtasks {
		where = append(where, "t.parent_id IS NULL")
	}

	tasks, err := queryTasks(ctx, s.db, strings.Join(where, " AND "), args, orderClause(f.Order), f.Limit)
	if err != nil {
		return nil, err
	}
	if f.WithSubtasks {
		if err := attachSubtasks(ctx, s.db, tasks); err != nil {
			return nil, err
		}
	}
	return tasks, nil
}

// attachSubtasks loads non-trashed children into each task's Subtasks.
func attachSubtasks(ctx context.Context, q querier, tasks []model.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	ids := make([]int64, len(tasks))
	index := make(map[int64]int, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
		index[t.ID] = i
	}
	marks, args := inClause(ids)
	children, err := queryTasks(ctx, q, "t.deleted_at IS NULL AND t.parent_id IN ("+marks+")", args, "t.position, t.id", 0)
	if err != nil {
		return err
	}
	for _, c := range children {
		i := index[c.ParentID]
		tasks[i].Subtasks = append(tasks[i].Subtasks, c)
	}
	return nil
}

// GetTask returns one non-trashed task with its subtasks.
func (s *Store) GetTask(ctx context.Context, id int64) (model.Task, error) {
	tasks, err := queryTasks(ctx, s.db, "t.id = ? AND t.deleted_at IS NULL", []any{id}, "t.id", 1)
	if err != nil {
		return model.Task{}, err
	}
	if len(tasks) == 0 {
		return model.Task{}, model.ErrNotFound
	}
	if err := attachSubtasks(ctx, s.db, tasks); err != nil {
		return model.Task{}, err
	}
	return tasks[0], nil
}

// CreateTask inserts a task (and any t.Subtasks) and returns the new id.
func (s *Store) CreateTask(ctx context.Context, t *model.Task) (int64, error) {
	var id int64
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var err error
		id, err = s.createTaskTx(ctx, tx, t)
		return err
	})
	return id, err
}

func (s *Store) createTaskTx(ctx context.Context, tx *sql.Tx, t *model.Task) (int64, error) {
	now := s.ts()
	var position int64
	if err := tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(position), 0) + 1 FROM tasks").Scan(&position); err != nil {
		return 0, err
	}
	res, err := tx.ExecContext(ctx, `
		INSERT INTO tasks (title, description, priority, due_date, project_id, parent_id, recurrence, position, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.Title, t.Description, t.Priority, nullStr(t.DueDate), nullID(t.ProjectID), nullID(t.ParentID),
		nullStr(t.Recurrence), position, now, now)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := setTags(ctx, tx, "task_tags", "task_id", id, t.Tags); err != nil {
		return 0, err
	}
	if err := indexTask(ctx, tx, id); err != nil {
		return 0, err
	}
	for i := range t.Subtasks {
		sub := t.Subtasks[i]
		sub.ParentID = id
		sub.ProjectID = t.ProjectID
		if _, err := s.createTaskTx(ctx, tx, &sub); err != nil {
			return 0, err
		}
	}
	return id, nil
}

// UpdateTask saves the editable fields of a task. Subtasks follow the task's project.
func (s *Store) UpdateTask(ctx context.Context, t *model.Task) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `
			UPDATE tasks SET title = ?, description = ?, priority = ?, due_date = ?, project_id = ?, recurrence = ?, updated_at = ?
			WHERE id = ? AND deleted_at IS NULL`,
			t.Title, t.Description, t.Priority, nullStr(t.DueDate), nullID(t.ProjectID), nullStr(t.Recurrence), s.ts(), t.ID)
		if err != nil {
			return err
		}
		if err := requireAffected(res); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "UPDATE tasks SET project_id = ? WHERE parent_id = ?", nullID(t.ProjectID), t.ID); err != nil {
			return err
		}
		if err := setTags(ctx, tx, "task_tags", "task_id", t.ID, t.Tags); err != nil {
			return err
		}
		return indexTask(ctx, tx, t.ID)
	})
}

// CompleteTask marks a task and its open subtasks done and dismisses its reminders.
// If next is not nil it is created in the same transaction and its id returned.
func (s *Store) CompleteTask(ctx context.Context, id int64, next *model.Task) (int64, error) {
	var nextID int64
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		now := s.ts()
		var exists int
		err := tx.QueryRowContext(ctx, "SELECT 1 FROM tasks WHERE id = ? AND deleted_at IS NULL", id).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrNotFound
		}
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE tasks SET status = 'done', completed_at = ?, updated_at = ?
			WHERE (id = ? OR parent_id = ?) AND status = 'todo' AND deleted_at IS NULL`, now, now, id, id); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			"UPDATE reminders SET dismissed_at = ? WHERE task_id = ? AND dismissed_at IS NULL", now, id); err != nil {
			return err
		}
		if next != nil {
			nextID, err = s.createTaskTx(ctx, tx, next)
		}
		return err
	})
	return nextID, err
}

// ReopenTask marks a task as not done.
func (s *Store) ReopenTask(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE tasks SET status = 'todo', completed_at = NULL, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL`, s.ts(), id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

// DeleteTask moves a task and its subtasks to the trash.
func (s *Store) DeleteTask(ctx context.Context, id int64) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		ids, err := idsWhere(ctx, tx,
			"SELECT id FROM tasks WHERE (id = ? OR parent_id = ?) AND deleted_at IS NULL", id, id)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return model.ErrNotFound
		}
		marks, args := inClause(ids)
		stamp := s.ts()
		if _, err := tx.ExecContext(ctx, "UPDATE tasks SET deleted_at = ? WHERE id IN ("+marks+")", append([]any{stamp}, args...)...); err != nil {
			return err
		}
		for _, tid := range ids {
			if err := removeIndex(ctx, tx, "task", tid); err != nil {
				return err
			}
		}
		return nil
	})
}

// ReorderTasks applies a manual order to the given task ids.
func (s *Store) ReorderTasks(ctx context.Context, ids []int64) error {
	return s.reorder(ctx, "tasks", ids)
}

// idsWhere runs a query returning a single id column.
func idsWhere(ctx context.Context, q querier, query string, args ...any) ([]int64, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

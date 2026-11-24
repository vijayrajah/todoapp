package store

import (
	"context"
	"database/sql"

	"todoapp/internal/model"
)

// setChecklist replaces the checklist of a task.
func setChecklist(ctx context.Context, tx *sql.Tx, taskID int64, items []model.ChecklistItem) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM task_checklist WHERE task_id = ?", taskID); err != nil {
		return err
	}
	for i, it := range items {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO task_checklist (task_id, text, done, position) VALUES (?, ?, ?, ?)",
			taskID, it.Text, boolInt(it.Done), i); err != nil {
			return err
		}
	}
	return nil
}

// loadChecklists returns checklist items per task id, in order.
func loadChecklists(ctx context.Context, q querier, ids []int64) (map[int64][]model.ChecklistItem, error) {
	result := make(map[int64][]model.ChecklistItem, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	marks, args := inClause(ids)
	rows, err := q.QueryContext(ctx,
		"SELECT task_id, id, text, done FROM task_checklist WHERE task_id IN ("+marks+") ORDER BY position, id", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var taskID int64
		var it model.ChecklistItem
		var done int
		if err := rows.Scan(&taskID, &it.ID, &it.Text, &done); err != nil {
			return nil, err
		}
		it.Done = done != 0
		result[taskID] = append(result[taskID], it)
	}
	return result, rows.Err()
}

package store

import (
	"context"

	"todoapp/internal/model"
)

const reminderSelect = `
	SELECT id, message, COALESCE(task_id, 0), COALESCE(note_id, 0), remind_at,
	  COALESCE(snoozed_until, ''), COALESCE(dismissed_at, ''), created_at
	FROM reminders`

func queryReminders(ctx context.Context, q querier, where string, args []any, orderBy string) ([]model.Reminder, error) {
	rows, err := q.QueryContext(ctx, reminderSelect+" WHERE "+where+" ORDER BY "+orderBy, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	reminders := []model.Reminder{}
	for rows.Next() {
		var r model.Reminder
		if err := rows.Scan(&r.ID, &r.Message, &r.TaskID, &r.NoteID, &r.RemindAt, &r.SnoozedUntil, &r.DismissedAt, &r.CreatedAt); err != nil {
			return nil, err
		}
		reminders = append(reminders, r)
	}
	return reminders, rows.Err()
}

const effectiveTime = "COALESCE(snoozed_until, remind_at)"

// ListReminders returns reminders that have not been dismissed, soonest first.
func (s *Store) ListReminders(ctx context.Context) ([]model.Reminder, error) {
	return queryReminders(ctx, s.db, "dismissed_at IS NULL", nil, effectiveTime+", id")
}

// DueReminders returns undismissed reminders whose effective time is at or before now.
func (s *Store) DueReminders(ctx context.Context, now string) ([]model.Reminder, error) {
	return queryReminders(ctx, s.db, "dismissed_at IS NULL AND "+effectiveTime+" <= ?", []any{now}, effectiveTime+", id")
}

// CreateReminder inserts a reminder and returns its id.
func (s *Store) CreateReminder(ctx context.Context, r *model.Reminder) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO reminders (message, task_id, note_id, remind_at, created_at) VALUES (?, ?, ?, ?, ?)",
		r.Message, nullID(r.TaskID), nullID(r.NoteID), r.RemindAt, s.ts())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetReminder returns one reminder.
func (s *Store) GetReminder(ctx context.Context, id int64) (model.Reminder, error) {
	reminders, err := queryReminders(ctx, s.db, "id = ?", []any{id}, "id")
	if err != nil {
		return model.Reminder{}, err
	}
	if len(reminders) == 0 {
		return model.Reminder{}, model.ErrNotFound
	}
	return reminders[0], nil
}

// SnoozeReminder postpones an undismissed reminder until the given timestamp.
func (s *Store) SnoozeReminder(ctx context.Context, id int64, until string) error {
	res, err := s.db.ExecContext(ctx, "UPDATE reminders SET snoozed_until = ? WHERE id = ? AND dismissed_at IS NULL", until, id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

// DismissReminder marks a reminder as handled.
func (s *Store) DismissReminder(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, "UPDATE reminders SET dismissed_at = ? WHERE id = ? AND dismissed_at IS NULL", s.ts(), id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

// DeleteReminder removes a reminder.
func (s *Store) DeleteReminder(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM reminders WHERE id = ?", id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

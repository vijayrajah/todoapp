package store

import (
	"context"

	"todoapp/internal/model"
)

// GetBackupSettings returns the single settings row.
func (s *Store) GetBackupSettings(ctx context.Context) (model.BackupSettings, error) {
	var b model.BackupSettings
	var enabled int
	err := s.db.QueryRowContext(ctx,
		"SELECT enabled, directory, interval_hours, keep, last_backup_at, last_error FROM backup_settings WHERE id = 1").
		Scan(&enabled, &b.Directory, &b.IntervalHours, &b.Keep, &b.LastBackupAt, &b.LastError)
	b.Enabled = enabled != 0
	return b, err
}

// UpdateBackupSettings saves the user-editable settings.
func (s *Store) UpdateBackupSettings(ctx context.Context, b model.BackupSettings) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE backup_settings SET enabled = ?, directory = ?, interval_hours = ?, keep = ? WHERE id = 1",
		boolInt(b.Enabled), b.Directory, b.IntervalHours, b.Keep)
	return err
}

// RecordBackupResult stores the outcome of the latest backup attempt.
// On success lastErr is empty and last_backup_at is set to now.
func (s *Store) RecordBackupResult(ctx context.Context, lastErr string) error {
	if lastErr == "" {
		_, err := s.db.ExecContext(ctx, "UPDATE backup_settings SET last_backup_at = ?, last_error = '' WHERE id = 1", s.ts())
		return err
	}
	_, err := s.db.ExecContext(ctx, "UPDATE backup_settings SET last_error = ? WHERE id = 1", lastErr)
	return err
}

package store

import (
	"context"

	"todoapp/internal/model"
)

// GetPreferences returns the single preferences row (digest_last_sent included).
func (s *Store) GetPreferences(ctx context.Context) (model.Preferences, error) {
	return getPreferences(ctx, s.db)
}

func getPreferences(ctx context.Context, q querier) (model.Preferences, error) {
	var p model.Preferences
	var enabled int
	err := q.QueryRowContext(ctx,
		"SELECT theme, accent, digest_enabled, digest_time, digest_webhook, digest_last_sent FROM preferences WHERE id = 1").
		Scan(&p.Theme, &p.Accent, &enabled, &p.DigestTime, &p.DigestWebhook, &p.DigestLastSent)
	p.DigestEnabled = enabled != 0
	return p, err
}

// UpdateAppearance saves theme and accent.
func (s *Store) UpdateAppearance(ctx context.Context, p model.Preferences) error {
	_, err := s.db.ExecContext(ctx, "UPDATE preferences SET theme = ?, accent = ? WHERE id = 1", p.Theme, p.Accent)
	return err
}

// UpdateDigest saves digest settings.
func (s *Store) UpdateDigest(ctx context.Context, p model.Preferences) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE preferences SET digest_enabled = ?, digest_time = ?, digest_webhook = ? WHERE id = 1",
		boolInt(p.DigestEnabled), p.DigestTime, p.DigestWebhook)
	return err
}

// DigestShownOn returns the date the in-app digest was last shown.
func (s *Store) DigestShownOn(ctx context.Context) (string, error) {
	var d string
	err := s.db.QueryRowContext(ctx, "SELECT digest_last_shown FROM preferences WHERE id = 1").Scan(&d)
	return d, err
}

// MarkDigestShown records that the in-app digest was shown on date.
func (s *Store) MarkDigestShown(ctx context.Context, date string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE preferences SET digest_last_shown = ? WHERE id = 1", date)
	return err
}

// MarkDigestSent records that the webhook digest was sent on date.
func (s *Store) MarkDigestSent(ctx context.Context, date string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE preferences SET digest_last_sent = ? WHERE id = 1", date)
	return err
}

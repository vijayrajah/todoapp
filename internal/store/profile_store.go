package store

import (
	"context"

	"todoapp/internal/model"
)

// GetProfile returns the single profile row.
func (s *Store) GetProfile(ctx context.Context) (model.Profile, error) {
	return getProfile(ctx, s.db)
}

func getProfile(ctx context.Context, q querier) (model.Profile, error) {
	var p model.Profile
	err := q.QueryRowContext(ctx, "SELECT name, email, timezone, bio, avatar_color FROM profile WHERE id = 1").
		Scan(&p.Name, &p.Email, &p.Timezone, &p.Bio, &p.AvatarColor)
	return p, err
}

// UpdateProfile saves the profile.
func (s *Store) UpdateProfile(ctx context.Context, p model.Profile) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE profile SET name = ?, email = ?, timezone = ?, bio = ?, avatar_color = ? WHERE id = 1",
		p.Name, p.Email, p.Timezone, p.Bio, p.AvatarColor)
	return err
}

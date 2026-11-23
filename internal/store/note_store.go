package store

import (
	"context"
	"database/sql"
	"strings"

	"todoapp/internal/model"
)

const noteSelect = `
	SELECT n.id, n.title, n.body, n.pinned, COALESCE(n.deleted_at, ''), n.created_at, n.updated_at
	FROM notes n`

func queryNotes(ctx context.Context, q querier, where string, args []any, orderBy string) ([]model.Note, error) {
	rows, err := q.QueryContext(ctx, noteSelect+" WHERE "+where+" ORDER BY "+orderBy, args...)
	if err != nil {
		return nil, err
	}
	notes := []model.Note{}
	for rows.Next() {
		var n model.Note
		var pinned int
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &pinned, &n.DeletedAt, &n.CreatedAt, &n.UpdatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		n.Pinned = pinned == 1
		n.Tags = []string{}
		notes = append(notes, n)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	ids := make([]int64, len(notes))
	for i, n := range notes {
		ids[i] = n.ID
	}
	tags, err := loadTags(ctx, q, "note_tags", "note_id", ids)
	if err != nil {
		return nil, err
	}
	for i := range notes {
		if tg := tags[notes[i].ID]; tg != nil {
			notes[i].Tags = tg
		}
	}
	return notes, nil
}

// ListNotes returns non-trashed notes, pinned first then most recently updated. tag is optional.
func (s *Store) ListNotes(ctx context.Context, tag string) ([]model.Note, error) {
	where := []string{"n.deleted_at IS NULL"}
	var args []any
	if tag != "" {
		where = append(where, `EXISTS (SELECT 1 FROM note_tags nt JOIN tags g ON g.id = nt.tag_id WHERE nt.note_id = n.id AND g.name = ?)`)
		args = append(args, tag)
	}
	return queryNotes(ctx, s.db, strings.Join(where, " AND "), args, "n.pinned DESC, n.updated_at DESC, n.id DESC")
}

// GetNote returns one non-trashed note.
func (s *Store) GetNote(ctx context.Context, id int64) (model.Note, error) {
	notes, err := queryNotes(ctx, s.db, "n.id = ? AND n.deleted_at IS NULL", []any{id}, "n.id")
	if err != nil {
		return model.Note{}, err
	}
	if len(notes) == 0 {
		return model.Note{}, model.ErrNotFound
	}
	return notes[0], nil
}

// CreateNote inserts a note and returns its id.
func (s *Store) CreateNote(ctx context.Context, n *model.Note) (int64, error) {
	var id int64
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		now := s.ts()
		res, err := tx.ExecContext(ctx,
			"INSERT INTO notes (title, body, pinned, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			n.Title, n.Body, boolInt(n.Pinned), now, now)
		if err != nil {
			return err
		}
		if id, err = res.LastInsertId(); err != nil {
			return err
		}
		if err := setTags(ctx, tx, "note_tags", "note_id", id, n.Tags); err != nil {
			return err
		}
		return indexNote(ctx, tx, id)
	})
	return id, err
}

// UpdateNote saves title, body, pinned flag and tags.
func (s *Store) UpdateNote(ctx context.Context, n *model.Note) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx,
			"UPDATE notes SET title = ?, body = ?, pinned = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
			n.Title, n.Body, boolInt(n.Pinned), s.ts(), n.ID)
		if err != nil {
			return err
		}
		if err := requireAffected(res); err != nil {
			return err
		}
		if err := setTags(ctx, tx, "note_tags", "note_id", n.ID, n.Tags); err != nil {
			return err
		}
		return indexNote(ctx, tx, n.ID)
	})
}

// DeleteNote moves a note to the trash.
func (s *Store) DeleteNote(ctx context.Context, id int64) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, "UPDATE notes SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL", s.ts(), id)
		if err != nil {
			return err
		}
		if err := requireAffected(res); err != nil {
			return err
		}
		return removeIndex(ctx, tx, "note", id)
	})
}

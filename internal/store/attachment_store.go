package store

import (
	"context"
	"database/sql"
	"errors"

	"todoapp/internal/model"
)

const attachmentMetaSelect = `SELECT id, COALESCE(task_id, 0), COALESCE(note_id, 0), filename, content_type, size, created_at FROM attachments`

func scanAttachment(row scanner) (model.Attachment, error) {
	var a model.Attachment
	err := row.Scan(&a.ID, &a.TaskID, &a.NoteID, &a.Filename, &a.ContentType, &a.Size, &a.CreatedAt)
	return a, err
}

// ownerColumn maps an owner kind ("task" or "note") to its column and table.
func ownerColumn(kind string) (col, table string, err error) {
	switch kind {
	case "task":
		return "task_id", "tasks", nil
	case "note":
		return "note_id", "notes", nil
	}
	return "", "", model.ValidationError("unknown attachment owner")
}

// ListAttachments returns attachment metadata (no content) of a task or note.
func (s *Store) ListAttachments(ctx context.Context, kind string, ownerID int64) ([]model.Attachment, error) {
	col, _, err := ownerColumn(kind)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, attachmentMetaSelect+" WHERE "+col+" = ? ORDER BY id", ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Attachment{}
	for rows.Next() {
		a, err := scanAttachment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// AddAttachment stores a file as a blob on a non-trashed task or note.
func (s *Store) AddAttachment(ctx context.Context, kind string, ownerID int64, filename, contentType string, data []byte) (model.Attachment, error) {
	col, table, err := ownerColumn(kind)
	if err != nil {
		return model.Attachment{}, err
	}
	var id int64
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		var exists int
		if err := tx.QueryRowContext(ctx, "SELECT 1 FROM "+table+" WHERE id = ? AND deleted_at IS NULL", ownerID).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
			return model.ErrNotFound
		} else if err != nil {
			return err
		}
		var count int
		if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM attachments WHERE "+col+" = ?", ownerID).Scan(&count); err != nil {
			return err
		}
		if count >= model.MaxAttachmentsPerOwner {
			return model.ValidationError("too many attachments")
		}
		res, err := tx.ExecContext(ctx,
			"INSERT INTO attachments ("+col+", filename, content_type, size, data, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			ownerID, filename, contentType, len(data), data, s.ts())
		if err != nil {
			return err
		}
		id, err = res.LastInsertId()
		return err
	})
	if err != nil {
		return model.Attachment{}, err
	}
	return scanAttachment(s.db.QueryRowContext(ctx, attachmentMetaSelect+" WHERE id = ?", id))
}

// GetAttachment returns an attachment with its content.
func (s *Store) GetAttachment(ctx context.Context, id int64) (model.AttachmentData, error) {
	var a model.AttachmentData
	err := s.db.QueryRowContext(ctx,
		"SELECT id, COALESCE(task_id, 0), COALESCE(note_id, 0), filename, content_type, size, created_at, data FROM attachments WHERE id = ?", id).
		Scan(&a.ID, &a.TaskID, &a.NoteID, &a.Filename, &a.ContentType, &a.Size, &a.CreatedAt, &a.Data)
	if errors.Is(err, sql.ErrNoRows) {
		return a, model.ErrNotFound
	}
	return a, err
}

// DeleteAttachment removes an attachment.
func (s *Store) DeleteAttachment(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM attachments WHERE id = ?", id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

func exportAttachments(ctx context.Context, q querier) ([]model.AttachmentData, error) {
	rows, err := q.QueryContext(ctx,
		"SELECT id, COALESCE(task_id, 0), COALESCE(note_id, 0), filename, content_type, size, created_at, data FROM attachments ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.AttachmentData{}
	for rows.Next() {
		var a model.AttachmentData
		if err := rows.Scan(&a.ID, &a.TaskID, &a.NoteID, &a.Filename, &a.ContentType, &a.Size, &a.CreatedAt, &a.Data); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

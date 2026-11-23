package store

import (
	"context"
	"database/sql"

	"todoapp/internal/model"
)

// setTags replaces the tags of one item. linkTable/idCol are trusted constants.
func setTags(ctx context.Context, tx *sql.Tx, linkTable, idCol string, id int64, tags []string) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM "+linkTable+" WHERE "+idCol+" = ?", id); err != nil {
		return err
	}
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO tags (name) VALUES (?)", tag); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			"INSERT OR IGNORE INTO "+linkTable+" ("+idCol+", tag_id) SELECT ?, id FROM tags WHERE name = ?", id, tag); err != nil {
			return err
		}
	}
	return nil
}

// loadTags returns tag names per item id. linkTable/idCol are trusted constants.
func loadTags(ctx context.Context, q querier, linkTable, idCol string, ids []int64) (map[int64][]string, error) {
	result := make(map[int64][]string, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	marks, args := inClause(ids)
	rows, err := q.QueryContext(ctx,
		"SELECT l."+idCol+", g.name FROM "+linkTable+" l JOIN tags g ON g.id = l.tag_id WHERE l."+idCol+" IN ("+marks+") ORDER BY g.name", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		result[id] = append(result[id], name)
	}
	return result, rows.Err()
}

// ListTags returns tags in use by non-trashed tasks or notes, with counts.
func (s *Store) ListTags(ctx context.Context) ([]model.Tag, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT g.name,
		  (SELECT COUNT(*) FROM task_tags tt JOIN tasks t ON t.id = tt.task_id
		     WHERE tt.tag_id = g.id AND t.deleted_at IS NULL),
		  (SELECT COUNT(*) FROM note_tags nt JOIN notes n ON n.id = nt.note_id
		     WHERE nt.tag_id = g.id AND n.deleted_at IS NULL)
		FROM tags g ORDER BY g.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags := []model.Tag{}
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.Name, &t.TaskCount, &t.NoteCount); err != nil {
			return nil, err
		}
		if t.TaskCount+t.NoteCount > 0 {
			tags = append(tags, t)
		}
	}
	return tags, rows.Err()
}

package store

import (
	"context"
	"database/sql"
	"html"
	"strings"
	"unicode"

	"todoapp/internal/model"
)

const (
	highlightStart = "\x01"
	highlightEnd   = "\x02"
)

// indexTask rebuilds the search entry for a task (removed when trashed).
func indexTask(ctx context.Context, tx *sql.Tx, id int64) error {
	if err := removeIndex(ctx, tx, "task", id); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO search_index (kind, ref_id, title, body, tags)
		SELECT 'task', t.id, t.title, t.description,
		  COALESCE((SELECT group_concat(g.name, ' ') FROM task_tags tt JOIN tags g ON g.id = tt.tag_id WHERE tt.task_id = t.id), '')
		FROM tasks t WHERE t.id = ? AND t.deleted_at IS NULL`, id)
	return err
}

// indexNote rebuilds the search entry for a note (removed when trashed).
func indexNote(ctx context.Context, tx *sql.Tx, id int64) error {
	if err := removeIndex(ctx, tx, "note", id); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO search_index (kind, ref_id, title, body, tags)
		SELECT 'note', n.id, n.title, n.body,
		  COALESCE((SELECT group_concat(g.name, ' ') FROM note_tags nt JOIN tags g ON g.id = nt.tag_id WHERE nt.note_id = n.id), '')
		FROM notes n WHERE n.id = ? AND n.deleted_at IS NULL`, id)
	return err
}

// indexProject rebuilds the search entry for a project (removed when trashed).
func indexProject(ctx context.Context, tx *sql.Tx, id int64) error {
	if err := removeIndex(ctx, tx, "project", id); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO search_index (kind, ref_id, title, body, tags)
		SELECT 'project', p.id, p.name, p.description, ''
		FROM projects p WHERE p.id = ? AND p.deleted_at IS NULL`, id)
	return err
}

func removeIndex(ctx context.Context, tx *sql.Tx, kind string, id int64) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM search_index WHERE kind = ? AND ref_id = ?", kind, id)
	return err
}

// ftsQuery turns free text into a safe FTS5 prefix query: every word must match.
func ftsQuery(text string) string {
	var terms []string
	for _, word := range strings.Fields(text) {
		clean := strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				return r
			}
			return -1
		}, word)
		if clean != "" {
			terms = append(terms, `"`+clean+`"*`)
		}
	}
	return strings.Join(terms, " ")
}

// Search runs a full-text query over tasks, notes and projects.
func (s *Store) Search(ctx context.Context, text string, limit int) ([]model.SearchResult, error) {
	results := []model.SearchResult{}
	match := ftsQuery(text)
	if match == "" {
		return results, nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT kind, ref_id, title, snippet(search_index, 3, ?, ?, '…', 16)
		FROM search_index WHERE search_index MATCH ? ORDER BY rank LIMIT ?`,
		highlightStart, highlightEnd, match, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var r model.SearchResult
		var snippet string
		if err := rows.Scan(&r.Kind, &r.ID, &r.Title, &snippet); err != nil {
			return nil, err
		}
		r.Snippet = safeSnippet(snippet)
		results = append(results, r)
	}
	return results, rows.Err()
}

// safeSnippet escapes the text and turns the highlight markers into <mark> tags.
func safeSnippet(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = html.EscapeString(s)
	s = strings.ReplaceAll(s, highlightStart, "<mark>")
	return strings.ReplaceAll(s, highlightEnd, "</mark>")
}

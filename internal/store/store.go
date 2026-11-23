// Package store holds all SQL. It contains no business rules.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"todoapp/internal/model"
)

// Store is the data access layer over SQLite.
type Store struct {
	db  *sql.DB
	now func() time.Time
}

// New returns a Store using db.
func New(db *sql.DB) *Store {
	return &Store{db: db, now: time.Now}
}

// querier is satisfied by *sql.DB and *sql.Tx.
type querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// scanner is satisfied by *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func (s *Store) ts() string { return model.FormatTime(s.now()) }

// Ping checks that the database is reachable.
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

// inTx runs fn in a transaction. Inside fn use only the tx: the pool holds a single connection.
func (s *Store) inTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullID(id int64) any {
	if id == 0 {
		return nil
	}
	return id
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// inClause returns "?,?,?" and the matching args for ids.
func inClause(ids []int64) (string, []any) {
	marks := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		marks[i] = "?"
		args[i] = id
	}
	return strings.Join(marks, ","), args
}

func requireAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// reorder gives the ids the existing set of positions, in the new order.
// table must be a trusted constant ("tasks" or "projects").
func (s *Store) reorder(ctx context.Context, table string, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		marks, args := inClause(ids)
		rows, err := tx.QueryContext(ctx, "SELECT position FROM "+table+" WHERE id IN ("+marks+")", args...)
		if err != nil {
			return err
		}
		var positions []int64
		for rows.Next() {
			var p int64
			if err := rows.Scan(&p); err != nil {
				rows.Close()
				return err
			}
			positions = append(positions, p)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		sort.Slice(positions, func(i, j int) bool { return positions[i] < positions[j] })
		if len(positions) != len(ids) {
			return model.ValidationError("unknown id in order list")
		}
		for i, id := range ids {
			if _, err := tx.ExecContext(ctx, "UPDATE "+table+" SET position = ? WHERE id = ?", positions[i], id); err != nil {
				return err
			}
		}
		return nil
	})
}

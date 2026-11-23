package service

import (
	"context"
	"path/filepath"
	"testing"

	"todoapp/internal/db"
	"todoapp/internal/model"
	"todoapp/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return store.New(conn)
}

func TestCompleteRecurringTaskCreatesNext(t *testing.T) {
	ctx := context.Background()
	tasks := NewTaskService(newTestStore(t))

	created, err := tasks.QuickAdd(ctx, "Pay rent 2026-01-31 every 1 month #bills", 0)
	if err != nil {
		t.Fatal(err)
	}
	done, next, err := tasks.Complete(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if done.Status != model.StatusDone {
		t.Errorf("status = %s, want done", done.Status)
	}
	if next == nil || next.DueDate != "2026-02-28" || len(next.Tags) != 1 {
		t.Fatalf("unexpected next occurrence: %+v", next)
	}
}

func TestOverviewAndSearch(t *testing.T) {
	ctx := context.Background()
	st := newTestStore(t)
	tasks := NewTaskService(st)
	overview := NewOverviewService(st, NewReminderService(st))

	if _, err := tasks.QuickAdd(ctx, "Urgent thing !high", 0); err != nil {
		t.Fatal(err)
	}
	if _, err := tasks.QuickAdd(ctx, "Old thing 2000-01-01", 0); err != nil {
		t.Fatal(err)
	}
	ov, err := overview.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(ov.HighPriority) != 1 || len(ov.Overdue) != 1 {
		t.Errorf("high=%d overdue=%d, want 1 and 1", len(ov.HighPriority), len(ov.Overdue))
	}

	res, err := st.Search(ctx, "urg", 10)
	if err != nil || len(res) != 1 {
		t.Fatalf("search: %v %v", res, err)
	}
}

func TestDeleteAndRestoreTask(t *testing.T) {
	ctx := context.Background()
	st := newTestStore(t)
	tasks := NewTaskService(st)

	task, err := tasks.QuickAdd(ctx, "Temp task", 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := tasks.Delete(ctx, task.ID); err != nil {
		t.Fatal(err)
	}
	if res, _ := st.Search(ctx, "temp", 10); len(res) != 0 {
		t.Error("trashed task must not be searchable")
	}
	if err := st.RestoreTask(ctx, task.ID); err != nil {
		t.Fatal(err)
	}
	if res, _ := st.Search(ctx, "temp", 10); len(res) != 1 {
		t.Error("restored task must be searchable again")
	}
}

func TestBackupRoundTrip(t *testing.T) {
	ctx := context.Background()
	src := newTestStore(t)
	if _, err := NewTaskService(src).QuickAdd(ctx, "Keep me #x", 0); err != nil {
		t.Fatal(err)
	}
	backup, err := src.ExportAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	dst := newTestStore(t)
	if err := dst.ImportAll(ctx, backup); err != nil {
		t.Fatal(err)
	}
	got, err := NewTaskService(dst).List(ctx, "all", 0, "")
	if err != nil || len(got) != 1 || got[0].Title != "Keep me" {
		t.Fatalf("after import: %+v %v", got, err)
	}
}

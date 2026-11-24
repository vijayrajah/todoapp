package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"todoapp/internal/model"
)

func TestBackupRunPrunesOldFiles(t *testing.T) {
	ctx := context.Background()
	st := newTestStore(t)
	dir := filepath.Join(t.TempDir(), "bk")
	svc := NewBackupService(st, dir, zerolog.Nop())

	if err := st.UpdateBackupSettings(ctx, model.BackupSettings{Enabled: true, IntervalHours: 1, Keep: 2}); err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 4; i++ {
		svc.now = func() time.Time { return start.Add(time.Duration(i) * time.Second) }
		if _, err := svc.RunNow(ctx); err != nil {
			t.Fatal(err)
		}
	}
	unrelated := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(unrelated, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := svc.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0].Name != "todoapp-backup-20260101-100003.db" {
		t.Fatalf("unexpected files: %+v", files)
	}
	if _, err := os.Stat(unrelated); err != nil {
		t.Fatal("unrelated file was removed")
	}
	if s, _ := st.GetBackupSettings(ctx); s.LastBackupAt == "" || s.LastError != "" {
		t.Fatalf("result not recorded: %+v", s)
	}
}

func TestBackupDue(t *testing.T) {
	svc := NewBackupService(newTestStore(t), t.TempDir(), zerolog.Nop())
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	s := model.BackupSettings{IntervalHours: 6}
	if !svc.due(s) {
		t.Error("never backed up should be due")
	}
	s.LastBackupAt = model.FormatTime(now.Add(-5 * time.Hour))
	if svc.due(s) {
		t.Error("5h ago with 6h interval should not be due")
	}
	s.LastBackupAt = model.FormatTime(now.Add(-7 * time.Hour))
	if !svc.due(s) {
		t.Error("7h ago with 6h interval should be due")
	}
}

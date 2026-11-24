package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"todoapp/internal/model"
	"todoapp/internal/store"
)

const (
	backupCheckEvery = time.Minute
	backupRetryAfter = 15 * time.Minute
)

// Only files with this exact shape are listed and pruned, so unrelated files in the folder are never touched.
var backupFileRe = regexp.MustCompile(`^todoapp-backup-\d{8}-\d{6}\.db$`)

// BackupService writes periodic SQLite snapshots and prunes old ones.
type BackupService struct {
	store      *store.Store
	defaultDir string
	log        zerolog.Logger
	now        func() time.Time

	mu          sync.Mutex // serialises runs
	lastAttempt time.Time
}

// NewBackupService returns a service that falls back to defaultDir when no location is configured.
func NewBackupService(st *store.Store, defaultDir string, log zerolog.Logger) *BackupService {
	return &BackupService{store: st, defaultDir: defaultDir, log: log, now: time.Now}
}

// Dir returns the effective backup directory for the settings.
func (b *BackupService) Dir(s model.BackupSettings) string {
	if s.Directory != "" {
		return s.Directory
	}
	return b.defaultDir
}

// Start runs the scheduler until ctx is cancelled.
func (b *BackupService) Start(ctx context.Context) {
	go func() {
		t := time.NewTicker(backupCheckEvery)
		defer t.Stop()
		b.tick(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				b.tick(ctx)
			}
		}
	}()
}

func (b *BackupService) tick(ctx context.Context) {
	s, err := b.store.GetBackupSettings(ctx)
	if err != nil {
		b.log.Error().Err(err).Msg("auto backup: read settings")
		return
	}
	if !s.Enabled || !b.due(s) {
		return
	}
	if _, err := b.RunNow(ctx); err != nil {
		b.log.Error().Err(err).Msg("auto backup failed")
	}
}

func (b *BackupService) due(s model.BackupSettings) bool {
	now := b.now()
	b.mu.Lock()
	lastAttempt := b.lastAttempt
	b.mu.Unlock()
	if !lastAttempt.IsZero() && s.LastError != "" && now.Sub(lastAttempt) < backupRetryAfter {
		return false
	}
	if s.LastBackupAt == "" {
		return true
	}
	last, err := time.Parse(model.TimeLayout, s.LastBackupAt)
	if err != nil {
		return true
	}
	return now.Sub(last) >= time.Duration(s.IntervalHours)*time.Hour
}

// RunNow creates a backup immediately, prunes old ones and records the result.
func (b *BackupService) RunNow(ctx context.Context) (model.BackupFile, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lastAttempt = b.now()

	s, err := b.store.GetBackupSettings(ctx)
	if err != nil {
		return model.BackupFile{}, err
	}
	file, err := b.write(ctx, s)
	if err != nil {
		if recErr := b.store.RecordBackupResult(ctx, err.Error()); recErr != nil {
			b.log.Error().Err(recErr).Msg("record backup result")
		}
		return model.BackupFile{}, err
	}
	if err := b.store.RecordBackupResult(ctx, ""); err != nil {
		return file, err
	}
	b.log.Info().Str("file", file.Name).Msg("backup created")
	return file, nil
}

func (b *BackupService) write(ctx context.Context, s model.BackupSettings) (model.BackupFile, error) {
	dir := b.Dir(s)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return model.BackupFile{}, fmt.Errorf("create backup folder: %w", err)
	}
	name := "todoapp-backup-" + b.now().UTC().Format("20060102-150405") + ".db"
	tmp := filepath.Join(dir, name+".tmp")
	defer os.Remove(tmp)
	if err := b.store.SnapshotTo(ctx, tmp); err != nil {
		return model.BackupFile{}, fmt.Errorf("write backup: %w", err)
	}
	final := filepath.Join(dir, name)
	if err := os.Rename(tmp, final); err != nil {
		return model.BackupFile{}, fmt.Errorf("finish backup: %w", err)
	}
	if err := prune(dir, s.Keep); err != nil {
		b.log.Warn().Err(err).Msg("prune old backups")
	}
	info, err := os.Stat(final)
	if err != nil {
		return model.BackupFile{}, err
	}
	return model.BackupFile{Name: name, Size: info.Size(), CreatedAt: model.FormatTime(info.ModTime())}, nil
}

// List returns existing backups, newest first.
func (b *BackupService) List(ctx context.Context) ([]model.BackupFile, error) {
	s, err := b.store.GetBackupSettings(ctx)
	if err != nil {
		return nil, err
	}
	names, err := backupNames(b.Dir(s))
	if err != nil {
		return nil, err
	}
	files := make([]model.BackupFile, 0, len(names))
	for i := len(names) - 1; i >= 0; i-- {
		info, err := os.Stat(filepath.Join(b.Dir(s), names[i]))
		if err != nil {
			continue
		}
		files = append(files, model.BackupFile{Name: names[i], Size: info.Size(), CreatedAt: model.FormatTime(info.ModTime())})
	}
	return files, nil
}

// backupNames returns matching file names, oldest first (names embed a sortable UTC timestamp).
func backupNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && backupFileRe.MatchString(e.Name()) {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func prune(dir string, keep int) error {
	names, err := backupNames(dir)
	if err != nil {
		return err
	}
	for len(names) > keep {
		if err := os.Remove(filepath.Join(dir, names[0])); err != nil {
			return err
		}
		names = names[1:]
	}
	return nil
}

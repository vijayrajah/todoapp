package model

import (
	"path/filepath"
	"strings"
)

// Limits for auto-backup settings.
const (
	MaxBackupIntervalHours = 24 * 365
	MaxBackupKeep          = 1000
)

// BackupSettings configures automatic database backups.
type BackupSettings struct {
	Enabled       bool   `json:"enabled"`
	Directory     string `json:"directory"` // "" = default directory under the data dir
	IntervalHours int    `json:"interval_hours"`
	Keep          int    `json:"keep"`
	LastBackupAt  string `json:"last_backup_at"` // read-only
	LastError     string `json:"last_error"`     // read-only
}

// Clean trims and validates user-editable fields.
func (b *BackupSettings) Clean() error {
	b.Directory = strings.TrimSpace(b.Directory)
	if len(b.Directory) > 500 || strings.ContainsRune(b.Directory, 0) {
		return ValidationError("backup location is invalid")
	}
	if b.Directory != "" && !filepath.IsAbs(b.Directory) {
		return ValidationError("backup location must be an absolute path")
	}
	if b.IntervalHours < 1 || b.IntervalHours > MaxBackupIntervalHours {
		return ValidationError("backup frequency must be between 1 and 8760 hours")
	}
	if b.Keep < 1 || b.Keep > MaxBackupKeep {
		return ValidationError("number of backups to keep must be between 1 and 1000")
	}
	return nil
}

// BackupFile describes one backup file on disk.
type BackupFile struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
}

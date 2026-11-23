package model

// BackupVersion is the current export document format.
const BackupVersion = 1

// Backup is the full export/import document.
type Backup struct {
	Version    int        `json:"version"`
	ExportedAt string     `json:"exported_at"`
	Profile    Profile    `json:"profile"`
	Projects   []Project  `json:"projects"`
	Tasks      []Task     `json:"tasks"` // flat, includes trashed
	Notes      []Note     `json:"notes"`
	Reminders  []Reminder `json:"reminders"`
}

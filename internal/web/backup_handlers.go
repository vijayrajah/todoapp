package web

import (
	"net/http"
	"os"
	"time"

	"todoapp/internal/model"
)

const maxImportBytes = 256 << 20

func (s *Server) exportBackup(w http.ResponseWriter, r *http.Request) {
	backup, err := s.store.ExportAll(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="todoapp-backup-`+time.Now().Format("2006-01-02")+`.json"`)
	writeJSON(w, http.StatusOK, backup)
}

func (s *Server) downloadDatabase(w http.ResponseWriter, r *http.Request) {
	tmp, err := os.CreateTemp("", "todoapp-snapshot-*.db")
	if err != nil {
		s.fail(w, r, err)
		return
	}
	path := tmp.Name()
	tmp.Close()
	defer os.Remove(path)

	if err := s.store.SnapshotTo(r.Context(), path); err != nil {
		s.fail(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="todoapp-`+time.Now().Format("2006-01-02")+`.db"`)
	http.ServeFile(w, r, path)
}

type backupSettingsResponse struct {
	model.BackupSettings
	EffectiveDirectory string `json:"effective_directory"`
}

func (s *Server) settingsResponse(r *http.Request) (backupSettingsResponse, error) {
	cur, err := s.store.GetBackupSettings(r.Context())
	if err != nil {
		return backupSettingsResponse{}, err
	}
	return backupSettingsResponse{BackupSettings: cur, EffectiveDirectory: s.backups.Dir(cur)}, nil
}

func (s *Server) getBackupSettings(w http.ResponseWriter, r *http.Request) {
	resp, err := s.settingsResponse(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) updateBackupSettings(w http.ResponseWriter, r *http.Request) {
	var b model.BackupSettings
	if err := readJSON(w, r, &b, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	if err := b.Clean(); err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.store.UpdateBackupSettings(r.Context(), b); err != nil {
		s.fail(w, r, err)
		return
	}
	s.getBackupSettings(w, r)
}

func (s *Server) listBackupFiles(w http.ResponseWriter, r *http.Request) {
	files, err := s.backups.List(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, files)
}

func (s *Server) runBackupNow(w http.ResponseWriter, r *http.Request) {
	file, err := s.backups.RunNow(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, file)
}

func (s *Server) importBackup(w http.ResponseWriter, r *http.Request) {
	var backup model.Backup
	if err := readJSON(w, r, &backup, maxImportBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.store.ImportAll(r.Context(), backup); err != nil {
		s.fail(w, r, err)
		return
	}
	s.log.Info().Int("tasks", len(backup.Tasks)).Int("notes", len(backup.Notes)).Msg("backup imported")
	w.WriteHeader(http.StatusNoContent)
}

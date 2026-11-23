package web

import (
	"net/http"
	"os"
	"time"

	"todoapp/internal/model"
)

const maxImportBytes = 64 << 20

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

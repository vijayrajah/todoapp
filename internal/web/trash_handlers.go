package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"todoapp/internal/model"
)

func (s *Server) listTrash(w http.ResponseWriter, r *http.Request) {
	trash, err := s.store.ListTrash(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, trash)
}

func (s *Server) restoreTrashed(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	ctx := r.Context()
	switch chi.URLParam(r, "kind") {
	case "task":
		err = s.store.RestoreTask(ctx, id)
	case "note":
		err = s.store.RestoreNote(ctx, id)
	case "project":
		err = s.store.RestoreProject(ctx, id)
	default:
		err = model.ValidationError("unknown kind")
	}
	if err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) purgeTrashed(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.store.PurgeTrashed(r.Context(), chi.URLParam(r, "kind"), id); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) emptyTrash(w http.ResponseWriter, r *http.Request) {
	if err := s.store.EmptyTrash(r.Context()); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

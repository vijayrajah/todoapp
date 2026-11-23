package web

import (
	"net/http"

	"todoapp/internal/model"
)

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.store.ListProjects(r.Context(), r.URL.Query().Get("archived") == "1")
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var p model.Project
	if err := readJSON(w, r, &p, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	if err := p.Clean(); err != nil {
		s.fail(w, r, err)
		return
	}
	id, err := s.store.CreateProject(r.Context(), &p)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	created, err := s.store.GetProject(r.Context(), id)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) reorderProjects(w http.ResponseWriter, r *http.Request) {
	var body idsBody
	if err := readJSON(w, r, &body, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.store.ReorderProjects(r.Context(), body.IDs); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getProject(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	p, err := s.store.GetProject(r.Context(), id)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) updateProject(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	var p model.Project
	if err := readJSON(w, r, &p, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	p.ID = id
	if err := p.Clean(); err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.store.UpdateProject(r.Context(), &p); err != nil {
		s.fail(w, r, err)
		return
	}
	updated, err := s.store.GetProject(r.Context(), id)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.store.DeleteProject(r.Context(), id); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) archiveProject(archived bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := pathID(r)
		if err != nil {
			s.fail(w, r, err)
			return
		}
		if err := s.store.SetProjectArchived(r.Context(), id, archived); err != nil {
			s.fail(w, r, err)
			return
		}
		p, err := s.store.GetProject(r.Context(), id)
		if err != nil {
			s.fail(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

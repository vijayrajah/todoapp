package web

import (
	"net/http"

	"todoapp/internal/model"
)

func (s *Server) listNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := s.notes.List(r.Context(), r.URL.Query().Get("tag"))
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, notes)
}

func (s *Server) createNote(w http.ResponseWriter, r *http.Request) {
	var n model.Note
	if err := readJSON(w, r, &n, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	created, err := s.notes.Create(r.Context(), n)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getNote(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	n, err := s.notes.Get(r.Context(), id)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (s *Server) updateNote(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	var n model.Note
	if err := readJSON(w, r, &n, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	n.ID = id
	updated, err := s.notes.Update(r.Context(), n)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteNote(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.store.DeleteNote(r.Context(), id); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

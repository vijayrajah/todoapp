package web

import (
	"net/http"

	"todoapp/internal/model"
)

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request) {
	p, err := s.store.GetProfile(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) updateProfile(w http.ResponseWriter, r *http.Request) {
	var p model.Profile
	if err := readJSON(w, r, &p, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	if err := p.Clean(); err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.store.UpdateProfile(r.Context(), p); err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

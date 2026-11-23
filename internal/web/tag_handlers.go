package web

import "net/http"

func (s *Server) listTags(w http.ResponseWriter, r *http.Request) {
	tags, err := s.store.ListTags(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, tags)
}

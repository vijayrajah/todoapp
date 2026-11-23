package web

import "net/http"

func (s *Server) overview(w http.ResponseWriter, r *http.Request) {
	o, err := s.overviewSvc.Get(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

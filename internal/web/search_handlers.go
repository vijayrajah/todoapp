package web

import "net/http"

const searchLimit = 50

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	results, err := s.store.Search(r.Context(), r.URL.Query().Get("q"), searchLimit)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

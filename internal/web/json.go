package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"todoapp/internal/model"
)

const maxBodyBytes = 1 << 20

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return // client went away; nothing useful to do
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// readJSON decodes the request body into v, limited to limit bytes.
func readJSON(w http.ResponseWriter, r *http.Request, v any, limit int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return model.ValidationError("invalid JSON body")
	}
	return nil
}

// fail maps an error to an HTTP response and logs unexpected ones.
func (s *Server) fail(w http.ResponseWriter, r *http.Request, err error) {
	var validation model.ValidationError
	switch {
	case errors.As(err, &validation):
		writeError(w, http.StatusBadRequest, validation.Error())
	case errors.Is(err, model.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	default:
		s.log.Error().Err(err).Str("method", r.Method).Str("path", r.URL.Path).Msg("request failed")
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

// pathID reads the {id} URL parameter.
func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, model.ValidationError("invalid id")
	}
	return id, nil
}

// queryID reads an optional numeric query parameter (0 when absent).
func queryID(r *http.Request, name string) (int64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return 0, nil
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 0 {
		return 0, model.ValidationError("invalid " + name)
	}
	return id, nil
}

// idsBody is the payload of reorder endpoints.
type idsBody struct {
	IDs []int64 `json:"ids"`
}

package web

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rs/zerolog"

	"todoapp/internal/buildinfo"
	"todoapp/internal/db"
	"todoapp/internal/store"
)

func testHandler(t *testing.T, readOnly bool) http.Handler {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	ui := fstest.MapFS{
		"templates/layout.html":         {Data: []byte(`{{define "layout"}}<h1>{{.Title}}</h1>{{template "content" .}}{{end}}`)},
		"templates/pages/overview.html": {Data: []byte(`{{define "content"}}ok{{end}}`)},
	}
	srv, err := NewServer(Options{
		Store: store.New(conn), UI: ui, Log: zerolog.Nop(), Build: buildinfo.New("t", "c"), ReadOnly: readOnly,
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv.Handler()
}

func do(h http.Handler, method, path, contentType, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealthAndPage(t *testing.T) {
	h := testHandler(t, false)
	if rec := do(h, "GET", "/healthz", "", ""); rec.Code != 200 {
		t.Errorf("healthz = %d", rec.Code)
	}
	rec := do(h, "GET", "/", "", "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Overview") {
		t.Errorf("overview page = %d %q", rec.Code, rec.Body.String())
	}
}

func TestWritesRequireJSON(t *testing.T) {
	h := testHandler(t, false)
	if rec := do(h, "POST", "/api/tasks", "text/plain", `{"title":"x"}`); rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("got %d, want 415", rec.Code)
	}
	if rec := do(h, "POST", "/api/tasks", "application/json", `{"title":"x"}`); rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
		t.Errorf("create = %d %s", rec.Code, rec.Body.String())
	}
}

func TestReadOnlyRejectsWrites(t *testing.T) {
	h := testHandler(t, true)
	if rec := do(h, "POST", "/api/tasks", "application/json", `{"title":"x"}`); rec.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", rec.Code)
	}
	if rec := do(h, "GET", "/api/tasks", "", ""); rec.Code != 200 {
		t.Errorf("read = %d", rec.Code)
	}
}

func TestValidationAndNotFound(t *testing.T) {
	h := testHandler(t, false)
	if rec := do(h, "POST", "/api/tasks", "application/json", `{"title":""}`); rec.Code != http.StatusBadRequest {
		t.Errorf("empty title = %d, want 400", rec.Code)
	}
	if rec := do(h, "GET", "/api/tasks/999", "", ""); rec.Code != http.StatusNotFound {
		t.Errorf("missing task = %d, want 404", rec.Code)
	}
}

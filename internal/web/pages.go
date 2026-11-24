package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"todoapp/internal/buildinfo"
	"todoapp/internal/model"
	"todoapp/internal/service"
)

// pageData is what every page template receives.
type pageData struct {
	Title     string
	Active    string
	View      string
	ProjectID int64
	Profile   model.Profile
	Today     string
	Build     buildinfo.Info
	ReadOnly  bool

	Theme         string
	AccentStyle   template.CSS
	Filters       []model.SavedFilter
	FilterID      int64
	DigestEnabled bool
}

type pageDef struct {
	path     string
	template string
	title    string
	active   string
	view     string
}

var pageDefs = []pageDef{
	{"/", "overview", "Overview", "overview", ""},
	{"/today", "tasks", "Today", "today", "today"},
	{"/upcoming", "tasks", "Upcoming", "upcoming", "upcoming"},
	{"/tasks", "tasks", "Tasks", "tasks", "all"},
	{"/completed", "tasks", "Completed", "tasks", "completed"},
	{"/calendar", "calendar", "Calendar", "calendar", ""},
	{"/templates", "templates", "Templates", "templates", ""},
	{"/views", "views", "Saved views", "views", ""},
	{"/projects", "projects", "Projects", "projects", ""},
	{"/notes", "notes", "Notes", "notes", ""},
	{"/reminders", "reminders", "Reminders", "reminders", ""},
	{"/tags", "tags", "Tags", "tags", ""},
	{"/search", "search", "Search", "", ""},
	{"/trash", "trash", "Trash", "trash", ""},
	{"/profile", "profile", "Profile", "profile", ""},
}

func (s *Server) registerPages(r chi.Router) {
	for _, def := range pageDefs {
		r.Get(def.path, s.page(def))
	}
	r.Get("/projects/{id}", s.projectPage)
	r.Get("/views/{id}", s.viewPage)
}

func (s *Server) page(def pageDef) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.renderPage(w, r, def, 0)
	}
}

func (s *Server) projectPage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}
	s.renderPage(w, r, pageDef{template: "project", title: "Project", active: "projects"}, id)
}

func (s *Server) viewPage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}
	f, err := s.store.GetFilter(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	s.renderPage(w, r, pageDef{template: "tasks", title: f.Name, active: "view-" + strconv.FormatInt(id, 10), view: "filter"}, id)
}

// accentStyle builds CSS custom properties from a validated #rrggbb colour (safe to emit as template.CSS).
func accentStyle(hex string) template.CSS {
	var r, g, b int
	if len(hex) != 7 || hex[0] != '#' {
		return ""
	}
	if _, err := fmt.Sscanf(hex[1:], "%02x%02x%02x", &r, &g, &b); err != nil {
		return ""
	}
	fg := "#ffffff"
	if (r*299+g*587+b*114)/1000 > 160 {
		fg = "#111111"
	}
	return template.CSS(fmt.Sprintf("--accent:%s;--accent-fg:%s;--accent-soft:rgba(%d,%d,%d,0.15);", hex, fg, r, g, b))
}

func (s *Server) renderPage(w http.ResponseWriter, r *http.Request, def pageDef, projectID int64) {
	profile, err := s.store.GetProfile(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	prefs, err := s.store.GetPreferences(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	filters, err := s.store.ListFilters(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	data := pageData{
		Theme: prefs.Theme, AccentStyle: accentStyle(prefs.Accent), Filters: filters, DigestEnabled: prefs.DigestEnabled,
		Title: def.title, Active: def.active, View: def.view, ProjectID: projectID, FilterID: projectID,
		Profile: profile, Today: service.TodayString(s.now(), profile), Build: s.build, ReadOnly: s.readOnly,
	}
	if err := s.renderer.Render(w, def.template, data); err != nil {
		s.log.Error().Err(err).Str("page", def.template).Msg("render page failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

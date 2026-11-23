package web

import (
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

func (s *Server) renderPage(w http.ResponseWriter, r *http.Request, def pageDef, projectID int64) {
	profile, err := s.store.GetProfile(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	data := pageData{
		Title: def.title, Active: def.active, View: def.view, ProjectID: projectID,
		Profile: profile, Today: service.TodayString(s.now(), profile), Build: s.build, ReadOnly: s.readOnly,
	}
	if err := s.renderer.Render(w, def.template, data); err != nil {
		s.log.Error().Err(err).Str("page", def.template).Msg("render page failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// Package web is the HTTP layer: pages, JSON API and middleware.
package web

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"todoapp/internal/buildinfo"
	"todoapp/internal/service"
	"todoapp/internal/store"
)

// Server wires handlers to services.
type Server struct {
	store       *store.Store
	tasks       *service.TaskService
	notes       *service.NoteService
	reminders   *service.ReminderService
	overviewSvc *service.OverviewService
	backups     *service.BackupService
	templates   *service.TemplateService
	filters     *service.FilterService
	attachments *service.AttachmentService
	digest      *service.DigestService
	renderer    *Renderer
	assets      fs.FS
	log         zerolog.Logger
	build       buildinfo.Info
	readOnly    bool
	now         func() time.Time
}

// Options configures NewServer.
type Options struct {
	Store    *store.Store
	UI       fs.FS // contains templates/ and static/
	Log      zerolog.Logger
	Build    buildinfo.Info
	ReadOnly bool
	Backups  *service.BackupService // optional; defaults to a service writing to ./backups
	Digest   *service.DigestService // optional; a default one is created
}

// NewServer builds the server and its services.
func NewServer(o Options) (*Server, error) {
	renderer, err := NewRenderer(o.UI)
	if err != nil {
		return nil, err
	}
	reminders := service.NewReminderService(o.Store)
	if o.Backups == nil {
		o.Backups = service.NewBackupService(o.Store, "backups", o.Log)
	}
	if o.Digest == nil {
		o.Digest = service.NewDigestService(o.Store, o.Log)
	}
	return &Server{
		backups:     o.Backups,
		templates:   service.NewTemplateService(o.Store),
		filters:     service.NewFilterService(o.Store),
		attachments: service.NewAttachmentService(o.Store),
		digest:      o.Digest,
		store:       o.Store,
		tasks:       service.NewTaskService(o.Store),
		notes:       service.NewNoteService(o.Store),
		reminders:   reminders,
		overviewSvc: service.NewOverviewService(o.Store, reminders),
		renderer:    renderer,
		assets:      o.UI,
		log:         o.Log,
		build:       o.Build,
		readOnly:    o.ReadOnly,
		now:         time.Now,
	}, nil
}

// Handler returns the full HTTP handler.
func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(recoverer(s.log), requestLogger(s.log))

	r.Get("/healthz", s.health)
	r.Handle("/static/*", noCache(http.FileServerFS(s.assets)))
	s.registerPages(r)

	r.Route("/api", func(api chi.Router) {
		api.Use(requireJSON)
		if s.readOnly {
			api.Use(readOnly)
		}
		s.registerAPI(api)
	})
	return r
}

func (s *Server) registerAPI(api chi.Router) {
	s.registerFeatureAPI(api)
	api.Get("/overview", s.overview)
	api.Get("/search", s.search)
	api.Get("/tags", s.listTags)
	api.Get("/profile", s.getProfile)
	api.Put("/profile", s.updateProfile)

	api.Route("/tasks", func(r chi.Router) {
		r.Get("/", s.listTasks)
		r.Post("/", s.createTask)
		r.Post("/quick", s.quickAddTask)
		r.Post("/reorder", s.reorderTasks)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", s.getTask)
			r.Put("/", s.updateTask)
			r.Delete("/", s.deleteTask)
			r.Post("/complete", s.completeTask)
			r.Post("/reopen", s.reopenTask)
		})
	})

	api.Route("/projects", func(r chi.Router) {
		r.Get("/", s.listProjects)
		r.Post("/", s.createProject)
		r.Post("/reorder", s.reorderProjects)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", s.getProject)
			r.Put("/", s.updateProject)
			r.Delete("/", s.deleteProject)
			r.Post("/archive", s.archiveProject(true))
			r.Post("/unarchive", s.archiveProject(false))
		})
	})

	api.Route("/notes", func(r chi.Router) {
		r.Get("/", s.listNotes)
		r.Post("/", s.createNote)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", s.getNote)
			r.Put("/", s.updateNote)
			r.Delete("/", s.deleteNote)
		})
	})

	api.Route("/reminders", func(r chi.Router) {
		r.Get("/", s.listReminders)
		r.Post("/", s.createReminder)
		r.Get("/due", s.dueReminders)
		r.Route("/{id}", func(r chi.Router) {
			r.Delete("/", s.deleteReminder)
			r.Post("/snooze", s.snoozeReminder)
			r.Post("/dismiss", s.dismissReminder)
		})
	})

	api.Route("/trash", func(r chi.Router) {
		r.Get("/", s.listTrash)
		r.Delete("/", s.emptyTrash)
		r.Post("/{kind}/{id}/restore", s.restoreTrashed)
		r.Delete("/{kind}/{id}", s.purgeTrashed)
	})

	api.Route("/backup", func(r chi.Router) {
		r.Get("/export", s.exportBackup)
		r.Get("/database", s.downloadDatabase)
		r.Post("/import", s.importBackup)
		r.Get("/settings", s.getBackupSettings)
		r.Put("/settings", s.updateBackupSettings)
		r.Get("/files", s.listBackupFiles)
		r.Post("/run", s.runBackupNow)
	})
}

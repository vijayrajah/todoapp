package web

import (
	"encoding/base64"
	"mime"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"todoapp/internal/model"
	"todoapp/internal/service"
)

// maxUploadBody covers a base64-encoded attachment plus JSON overhead.
const maxUploadBody = model.MaxAttachmentBytes*4/3 + 4096

// ---- templates ----

func (s *Server) listTemplates(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListTemplates(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) createTemplate(w http.ResponseWriter, r *http.Request) {
	var t model.Template
	if err := readJSON(w, r, &t, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	created, err := s.templates.Create(r.Context(), t)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) updateTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	var t model.Template
	if err := readJSON(w, r, &t, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	t.ID = id
	updated, err := s.templates.Update(r.Context(), t)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.store.DeleteTemplate(r.Context(), id); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) useTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	var in service.UseInput
	if err := readJSON(w, r, &in, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	t, err := s.templates.Use(r.Context(), id, in)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

// ---- saved filters ----

func (s *Server) listFilters(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListFilters(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) createFilter(w http.ResponseWriter, r *http.Request) {
	var f model.SavedFilter
	if err := readJSON(w, r, &f, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	created, err := s.filters.Create(r.Context(), f)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) updateFilter(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	var f model.SavedFilter
	if err := readJSON(w, r, &f, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	f.ID = id
	updated, err := s.filters.Update(r.Context(), f)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteFilter(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.store.DeleteFilter(r.Context(), id); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) filterTasks(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	tasks, err := s.filters.Tasks(r.Context(), id)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

// ---- attachments ----

func (s *Server) listAttachments(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := pathID(r)
		if err != nil {
			s.fail(w, r, err)
			return
		}
		list, err := s.store.ListAttachments(r.Context(), kind, id)
		if err != nil {
			s.fail(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, list)
	}
}

func (s *Server) addAttachment(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := pathID(r)
		if err != nil {
			s.fail(w, r, err)
			return
		}
		var body struct {
			Filename string `json:"filename"`
			Data     string `json:"data"` // base64
		}
		if err := readJSON(w, r, &body, maxUploadBody); err != nil {
			s.fail(w, r, model.ValidationError("upload is invalid or too large (max 10 MB)"))
			return
		}
		data, err := base64.StdEncoding.DecodeString(body.Data)
		if err != nil {
			s.fail(w, r, model.ValidationError("file data must be base64"))
			return
		}
		a, err := s.attachments.Add(r.Context(), kind, id, body.Filename, data)
		if err != nil {
			s.fail(w, r, err)
			return
		}
		writeJSON(w, http.StatusCreated, a)
	}
}

// inlineTypes are the only types served inline; everything else is a download.
var inlineTypes = map[string]bool{"image/png": true, "image/jpeg": true, "image/gif": true, "image/webp": true}

func (s *Server) serveAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	a, err := s.store.GetAttachment(r.Context(), id)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	h := w.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Content-Security-Policy", "default-src 'none'; sandbox")
	h.Set("Content-Length", strconv.Itoa(len(a.Data)))
	if inlineTypes[a.ContentType] && r.URL.Query().Get("download") == "" {
		h.Set("Content-Type", a.ContentType)
		h.Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": a.Filename}))
	} else {
		h.Set("Content-Type", "application/octet-stream")
		h.Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": a.Filename}))
	}
	if _, err := w.Write(a.Data); err != nil {
		return // client went away
	}
}

func (s *Server) deleteAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.store.DeleteAttachment(r.Context(), id); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- preferences ----

func (s *Server) getPreferences(w http.ResponseWriter, r *http.Request) {
	p, err := s.store.GetPreferences(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) updateAppearance(w http.ResponseWriter, r *http.Request) {
	var p model.Preferences
	if err := readJSON(w, r, &p, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	if err := p.CleanAppearance(); err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.store.UpdateAppearance(r.Context(), p); err != nil {
		s.fail(w, r, err)
		return
	}
	s.getPreferences(w, r)
}

func (s *Server) updateDigest(w http.ResponseWriter, r *http.Request) {
	var p model.Preferences
	if err := readJSON(w, r, &p, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	if err := p.CleanDigest(); err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.store.UpdateDigest(r.Context(), p); err != nil {
		s.fail(w, r, err)
		return
	}
	s.getPreferences(w, r)
}

// ---- digest ----

func (s *Server) getDigest(w http.ResponseWriter, r *http.Request) {
	d, err := s.digest.Current(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (s *Server) digestSeen(w http.ResponseWriter, r *http.Request) {
	if err := s.digest.MarkShown(r.Context()); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) digestSend(w http.ResponseWriter, r *http.Request) {
	if err := s.digest.SendNow(r.Context()); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) registerFeatureAPI(api chi.Router) {
	api.Route("/templates", func(r chi.Router) {
		r.Get("/", s.listTemplates)
		r.Post("/", s.createTemplate)
		r.Route("/{id}", func(r chi.Router) {
			r.Put("/", s.updateTemplate)
			r.Delete("/", s.deleteTemplate)
			r.Post("/use", s.useTemplate)
		})
	})
	api.Route("/filters", func(r chi.Router) {
		r.Get("/", s.listFilters)
		r.Post("/", s.createFilter)
		r.Route("/{id}", func(r chi.Router) {
			r.Put("/", s.updateFilter)
			r.Delete("/", s.deleteFilter)
			r.Get("/tasks", s.filterTasks)
		})
	})
	api.Get("/tasks/{id}/attachments", s.listAttachments("task"))
	api.Post("/tasks/{id}/attachments", s.addAttachment("task"))
	api.Get("/notes/{id}/attachments", s.listAttachments("note"))
	api.Post("/notes/{id}/attachments", s.addAttachment("note"))
	api.Get("/attachments/{id}", s.serveAttachment)
	api.Delete("/attachments/{id}", s.deleteAttachment)

	api.Get("/preferences", s.getPreferences)
	api.Put("/preferences/appearance", s.updateAppearance)
	api.Put("/preferences/digest", s.updateDigest)
	api.Get("/digest", s.getDigest)
	api.Post("/digest/seen", s.digestSeen)
	api.Post("/digest/send", s.digestSend)
}

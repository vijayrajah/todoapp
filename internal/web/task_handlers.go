package web

import (
	"net/http"

	"todoapp/internal/model"
)

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	projectID, err := queryID(r, "project")
	if err != nil {
		s.fail(w, r, err)
		return
	}
	q := r.URL.Query()
	var tasks []model.Task
	if q.Get("view") == "range" {
		tasks, err = s.tasks.ListRange(r.Context(), q.Get("from"), q.Get("to"))
	} else {
		tasks, err = s.tasks.List(r.Context(), q.Get("view"), projectID, q.Get("tag"))
	}
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var t model.Task
	if err := readJSON(w, r, &t, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	created, err := s.tasks.Create(r.Context(), t)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) quickAddTask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Text      string `json:"text"`
		ProjectID int64  `json:"project_id"`
	}
	if err := readJSON(w, r, &body, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	created, err := s.tasks.QuickAdd(r.Context(), body.Text, body.ProjectID)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) reorderTasks(w http.ResponseWriter, r *http.Request) {
	var body idsBody
	if err := readJSON(w, r, &body, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.tasks.Reorder(r.Context(), body.IDs); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	t, err := s.tasks.Get(r.Context(), id)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) updateTask(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	var t model.Task
	if err := readJSON(w, r, &t, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	t.ID = id
	updated, err := s.tasks.Update(r.Context(), t)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.tasks.Delete(r.Context(), id); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) completeTask(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	task, next, err := s.tasks.Complete(r.Context(), id)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"task": task, "next": next})
}

func (s *Server) reopenTask(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	t, err := s.tasks.Reopen(r.Context(), id)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

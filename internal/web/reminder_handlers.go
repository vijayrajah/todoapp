package web

import (
	"net/http"

	"todoapp/internal/model"
)

func (s *Server) listReminders(w http.ResponseWriter, r *http.Request) {
	reminders, err := s.store.ListReminders(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, reminders)
}

func (s *Server) dueReminders(w http.ResponseWriter, r *http.Request) {
	reminders, err := s.reminders.Due(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, reminders)
}

func (s *Server) createReminder(w http.ResponseWriter, r *http.Request) {
	var rem model.Reminder
	if err := readJSON(w, r, &rem, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	created, err := s.reminders.Create(r.Context(), rem)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) snoozeReminder(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	var body struct {
		Minutes int    `json:"minutes"`
		Until   string `json:"until"`
	}
	if err := readJSON(w, r, &body, maxBodyBytes); err != nil {
		s.fail(w, r, err)
		return
	}
	rem, err := s.reminders.Snooze(r.Context(), id, body.Minutes, body.Until)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, rem)
}

func (s *Server) dismissReminder(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.store.DismissReminder(r.Context(), id); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) deleteReminder(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if err := s.store.DeleteReminder(r.Context(), id); err != nil {
		s.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

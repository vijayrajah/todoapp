package service

import (
	"testing"
	"time"

	"todoapp/internal/model"
)

func TestWeekdayRecurrence(t *testing.T) {
	r, err := model.ParseRecurrence("weekly:1:mon,wed,fri")
	if err != nil {
		t.Fatal(err)
	}
	// 2024-01-01 is a Monday.
	cases := []struct{ from, want string }{
		{"2024-01-01", "2024-01-03"},
		{"2024-01-03", "2024-01-05"},
		{"2024-01-05", "2024-01-08"},
		{"2024-01-06", "2024-01-08"},
	}
	for _, c := range cases {
		got, err := NextDue(r, c.from, c.from)
		if err != nil || got != c.want {
			t.Errorf("NextDue(%s) = %s, %v, want %s", c.from, got, err, c.want)
		}
	}
	if _, err := model.ParseRecurrence("daily:1:mon"); err == nil {
		t.Error("days on a daily rule must be rejected")
	}
}

func TestQuickAddWeekdays(t *testing.T) {
	st := newTestStore(t)
	tasks := NewTaskService(st)
	task, err := tasks.QuickAdd(t.Context(), "Gym every mon wed fri", 0)
	if err != nil {
		t.Fatal(err)
	}
	if task.Recurrence != "weekly:1:mon,wed,fri" || task.DueDate == "" {
		t.Errorf("got %q due %q", task.Recurrence, task.DueDate)
	}
}

func TestToTaskFilter(t *testing.T) {
	wed := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	f := ToTaskFilter(model.SavedFilter{Due: model.DueThisWeek, MinPriority: 3, ProjectID: 2}, wed)
	if f.DueFrom != "2024-01-01" || f.DueTo != "2024-01-07" || f.MinPriority != 3 || f.ProjectID != 2 || f.Status != model.StatusTodo {
		t.Errorf("filter = %+v", f)
	}
}

func TestRecurringCopiesChecklistUndone(t *testing.T) {
	st := newTestStore(t)
	tasks := NewTaskService(st)
	created, err := tasks.Create(t.Context(), model.Task{
		Title: "Water", DueDate: "2030-01-01", Recurrence: "daily:1",
		Checklist: []model.ChecklistItem{{Text: "Ferns", Done: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, next, err := tasks.Complete(t.Context(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if next == nil || len(next.Checklist) != 1 || next.Checklist[0].Done {
		t.Errorf("next = %+v", next)
	}
}

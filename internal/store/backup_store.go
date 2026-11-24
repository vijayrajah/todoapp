package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"todoapp/internal/model"
)

// ExportAll returns every row, including trashed items.
func (s *Store) ExportAll(ctx context.Context) (model.Backup, error) {
	profile, err := s.GetProfile(ctx)
	if err != nil {
		return model.Backup{}, err
	}
	projects, err := queryProjects(ctx, s.db, "1 = 1", nil, "p.id")
	if err != nil {
		return model.Backup{}, err
	}
	tasks, err := queryTasks(ctx, s.db, "1 = 1", nil, "t.id", 0)
	if err != nil {
		return model.Backup{}, err
	}
	notes, err := queryNotes(ctx, s.db, "1 = 1", nil, "n.id")
	if err != nil {
		return model.Backup{}, err
	}
	reminders, err := queryReminders(ctx, s.db, "1 = 1", nil, "id")
	if err != nil {
		return model.Backup{}, err
	}
	templates, err := s.ListTemplates(ctx)
	if err != nil {
		return model.Backup{}, err
	}
	filters, err := s.ListFilters(ctx)
	if err != nil {
		return model.Backup{}, err
	}
	attachments, err := exportAttachments(ctx, s.db)
	if err != nil {
		return model.Backup{}, err
	}
	prefs, err := s.GetPreferences(ctx)
	if err != nil {
		return model.Backup{}, err
	}
	return model.Backup{
		Version:     model.BackupVersion,
		Templates:   templates,
		Filters:     filters,
		Attachments: attachments,
		Preferences: &prefs,
		ExportedAt:  s.ts(),
		Profile:     profile,
		Projects:    projects,
		Tasks:       tasks,
		Notes:       notes,
		Reminders:   reminders,
	}, nil
}

// ImportAll replaces all data with the backup contents in one transaction.
func (s *Store) ImportAll(ctx context.Context, b model.Backup) error {
	if b.Version != model.BackupVersion {
		return model.ValidationError(fmt.Sprintf("unsupported backup version %d", b.Version))
	}
	if err := b.Profile.Clean(); err != nil {
		return err
	}
	for i := range b.Projects {
		if err := b.Projects[i].Clean(); err != nil {
			return fmt.Errorf("project %d: %w", b.Projects[i].ID, err)
		}
	}
	for i := range b.Tasks {
		if err := b.Tasks[i].Clean(); err != nil {
			return fmt.Errorf("task %d: %w", b.Tasks[i].ID, err)
		}
	}
	for i := range b.Notes {
		if err := b.Notes[i].Clean(); err != nil {
			return fmt.Errorf("note %d: %w", b.Notes[i].ID, err)
		}
	}
	for i := range b.Templates {
		if err := b.Templates[i].Clean(); err != nil {
			return fmt.Errorf("template %d: %w", b.Templates[i].ID, err)
		}
	}
	for i := range b.Filters {
		if err := b.Filters[i].Clean(); err != nil {
			return fmt.Errorf("filter %d: %w", b.Filters[i].ID, err)
		}
	}
	for _, a := range b.Attachments {
		if len(a.Data) > model.MaxAttachmentBytes {
			return model.ValidationError("attachment is too large")
		}
	}
	if b.Preferences != nil {
		if err := b.Preferences.CleanAppearance(); err != nil {
			return err
		}
		if err := b.Preferences.CleanDigest(); err != nil {
			return err
		}
	}

	return s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, "PRAGMA defer_foreign_keys = ON"); err != nil {
			return err
		}
		for _, table := range []string{"attachments", "task_checklist", "task_templates", "saved_filters", "reminders", "task_tags", "note_tags", "tasks", "notes", "projects", "tags", "search_index", "sqlite_sequence"} {
			if _, err := tx.ExecContext(ctx, "DELETE FROM "+table); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx,
			"UPDATE profile SET name = ?, email = ?, timezone = ?, bio = ?, avatar_color = ? WHERE id = 1",
			b.Profile.Name, b.Profile.Email, b.Profile.Timezone, b.Profile.Bio, b.Profile.AvatarColor); err != nil {
			return err
		}
		if err := s.insertBackupRows(ctx, tx, b); err != nil {
			return err
		}
		return s.insertBackupExtras(ctx, tx, b)
	})
}

func (s *Store) insertBackupRows(ctx context.Context, tx *sql.Tx, b model.Backup) error {
	for _, p := range b.Projects {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO projects (id, name, description, color, position, archived, deleted_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			p.ID, p.Name, p.Description, p.Color, p.Position, boolInt(p.Archived), nullStr(p.DeletedAt), p.CreatedAt); err != nil {
			return fmt.Errorf("insert project %d: %w", p.ID, err)
		}
		if err := indexProject(ctx, tx, p.ID); err != nil {
			return err
		}
	}
	for _, t := range b.Tasks {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO tasks (id, title, description, status, priority, due_date, project_id, parent_id, recurrence,
			                   position, completed_at, deleted_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			t.ID, t.Title, t.Description, t.Status, t.Priority, nullStr(t.DueDate), nullID(t.ProjectID), nullID(t.ParentID),
			nullStr(t.Recurrence), t.Position, nullStr(t.CompletedAt), nullStr(t.DeletedAt), t.CreatedAt, t.UpdatedAt); err != nil {
			return fmt.Errorf("insert task %d: %w", t.ID, err)
		}
		if err := setTags(ctx, tx, "task_tags", "task_id", t.ID, t.Tags); err != nil {
			return err
		}
		if err := setChecklist(ctx, tx, t.ID, t.Checklist); err != nil {
			return err
		}
		if err := indexTask(ctx, tx, t.ID); err != nil {
			return err
		}
	}
	for _, n := range b.Notes {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO notes (id, title, body, pinned, deleted_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
			n.ID, n.Title, n.Body, boolInt(n.Pinned), nullStr(n.DeletedAt), n.CreatedAt, n.UpdatedAt); err != nil {
			return fmt.Errorf("insert note %d: %w", n.ID, err)
		}
		if err := setTags(ctx, tx, "note_tags", "note_id", n.ID, n.Tags); err != nil {
			return err
		}
		if err := indexNote(ctx, tx, n.ID); err != nil {
			return err
		}
	}
	for _, r := range b.Reminders {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO reminders (id, message, task_id, note_id, remind_at, snoozed_until, dismissed_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			r.ID, r.Message, nullID(r.TaskID), nullID(r.NoteID), r.RemindAt, nullStr(r.SnoozedUntil), nullStr(r.DismissedAt), r.CreatedAt); err != nil {
			return fmt.Errorf("insert reminder %d: %w", r.ID, err)
		}
	}
	return nil
}

// insertBackupExtras restores templates, saved filters, attachments and preferences.
func (s *Store) insertBackupExtras(ctx context.Context, tx *sql.Tx, b model.Backup) error {
	for _, t := range b.Templates {
		data, err := templateJSON(t)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO task_templates (id, name, data, position, created_at) VALUES (?, ?, ?, ?, ?)",
			t.ID, t.Name, data, t.Position, t.CreatedAt); err != nil {
			return fmt.Errorf("insert template %d: %w", t.ID, err)
		}
	}
	for _, f := range b.Filters {
		data, err := filterJSON(f)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO saved_filters (id, name, data, position, created_at) VALUES (?, ?, ?, ?, ?)",
			f.ID, f.Name, data, f.Position, f.CreatedAt); err != nil {
			return fmt.Errorf("insert filter %d: %w", f.ID, err)
		}
	}
	for _, a := range b.Attachments {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO attachments (id, task_id, note_id, filename, content_type, size, data, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			a.ID, nullID(a.TaskID), nullID(a.NoteID), model.CleanFilename(a.Filename), a.ContentType, len(a.Data), a.Data, a.CreatedAt); err != nil {
			return fmt.Errorf("insert attachment %d: %w", a.ID, err)
		}
	}
	if p := b.Preferences; p != nil {
		if _, err := tx.ExecContext(ctx,
			"UPDATE preferences SET theme = ?, accent = ?, digest_enabled = ?, digest_time = ?, digest_webhook = ? WHERE id = 1",
			p.Theme, p.Accent, boolInt(p.DigestEnabled), p.DigestTime, p.DigestWebhook); err != nil {
			return err
		}
	}
	return nil
}

// SnapshotTo writes a consistent copy of the database file to path.
func (s *Store) SnapshotTo(ctx context.Context, path string) error {
	// VACUUM INTO needs a non-existing target.
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	_, err := s.db.ExecContext(ctx, "VACUUM INTO ?", path)
	return err
}

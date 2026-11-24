CREATE TABLE task_checklist (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id  INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    text     TEXT NOT NULL,
    done     INTEGER NOT NULL DEFAULT 0,
    position INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX idx_checklist_task ON task_checklist(task_id);

CREATE TABLE task_templates (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    data       TEXT NOT NULL,
    position   INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL
);

CREATE TABLE saved_filters (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    data       TEXT NOT NULL,
    position   INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL
);

CREATE TABLE attachments (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id      INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
    note_id      INTEGER REFERENCES notes(id) ON DELETE CASCADE,
    filename     TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size         INTEGER NOT NULL,
    data         BLOB NOT NULL,
    created_at   TEXT NOT NULL,
    CHECK ((task_id IS NULL) <> (note_id IS NULL))
);
CREATE INDEX idx_attachments_task ON attachments(task_id);
CREATE INDEX idx_attachments_note ON attachments(note_id);

CREATE TABLE preferences (
    id                INTEGER PRIMARY KEY CHECK (id = 1),
    theme             TEXT NOT NULL DEFAULT 'auto',
    accent            TEXT NOT NULL DEFAULT '',
    digest_enabled    INTEGER NOT NULL DEFAULT 0,
    digest_time       TEXT NOT NULL DEFAULT '08:00',
    digest_webhook    TEXT NOT NULL DEFAULT '',
    digest_last_shown TEXT NOT NULL DEFAULT '',
    digest_last_sent  TEXT NOT NULL DEFAULT ''
);
INSERT INTO preferences (id) VALUES (1);

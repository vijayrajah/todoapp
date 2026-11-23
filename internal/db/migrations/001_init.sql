CREATE TABLE projects (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    color       TEXT NOT NULL DEFAULT '#5b6cf0',
    position    INTEGER NOT NULL DEFAULT 0,
    archived    INTEGER NOT NULL DEFAULT 0,
    deleted_at  TEXT,
    created_at  TEXT NOT NULL
);

CREATE TABLE tasks (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'todo' CHECK (status IN ('todo', 'done')),
    priority     INTEGER NOT NULL DEFAULT 0 CHECK (priority BETWEEN 0 AND 3),
    due_date     TEXT,
    project_id   INTEGER REFERENCES projects(id) ON DELETE CASCADE,
    parent_id    INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
    recurrence   TEXT,
    position     INTEGER NOT NULL DEFAULT 0,
    completed_at TEXT,
    deleted_at   TEXT,
    created_at   TEXT NOT NULL,
    updated_at   TEXT NOT NULL
);
CREATE INDEX idx_tasks_due ON tasks(due_date);
CREATE INDEX idx_tasks_project ON tasks(project_id);
CREATE INDEX idx_tasks_parent ON tasks(parent_id);

CREATE TABLE tags (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE task_tags (
    task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    tag_id  INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, tag_id)
);

CREATE TABLE notes (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    title      TEXT NOT NULL,
    body       TEXT NOT NULL DEFAULT '',
    pinned     INTEGER NOT NULL DEFAULT 0,
    deleted_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE note_tags (
    note_id INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    tag_id  INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (note_id, tag_id)
);

CREATE TABLE reminders (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    message       TEXT NOT NULL,
    task_id       INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
    note_id       INTEGER REFERENCES notes(id) ON DELETE CASCADE,
    remind_at     TEXT NOT NULL,
    snoozed_until TEXT,
    dismissed_at  TEXT,
    created_at    TEXT NOT NULL
);

CREATE TABLE profile (
    id           INTEGER PRIMARY KEY CHECK (id = 1),
    name         TEXT NOT NULL DEFAULT '',
    email        TEXT NOT NULL DEFAULT '',
    timezone     TEXT NOT NULL DEFAULT 'UTC',
    bio          TEXT NOT NULL DEFAULT '',
    avatar_color TEXT NOT NULL DEFAULT '#5b6cf0'
);
INSERT INTO profile (id) VALUES (1);

CREATE VIRTUAL TABLE search_index USING fts5(
    kind UNINDEXED,
    ref_id UNINDEXED,
    title,
    body,
    tags,
    tokenize = 'porter unicode61'
);

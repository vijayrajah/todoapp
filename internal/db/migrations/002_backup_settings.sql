CREATE TABLE backup_settings (
    id             INTEGER PRIMARY KEY CHECK (id = 1),
    enabled        INTEGER NOT NULL DEFAULT 0,
    directory      TEXT NOT NULL DEFAULT '',
    interval_hours INTEGER NOT NULL DEFAULT 24,
    keep           INTEGER NOT NULL DEFAULT 7,
    last_backup_at TEXT NOT NULL DEFAULT '',
    last_error     TEXT NOT NULL DEFAULT ''
);
INSERT INTO backup_settings (id) VALUES (1);

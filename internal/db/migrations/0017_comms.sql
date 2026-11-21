CREATE TABLE IF NOT EXISTS announcements (
  id         TEXT PRIMARY KEY,
  event_id   TEXT NOT NULL,
  title      TEXT NOT NULL,
  body       TEXT NOT NULL DEFAULT '',
  channels   TEXT NOT NULL DEFAULT 'push',
  audience   TEXT NOT NULL DEFAULT 'all',
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

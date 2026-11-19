-- Kata performances with panel judging.

CREATE TABLE IF NOT EXISTS kata_performances (
  id          TEXT PRIMARY KEY,
  event_id    TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  fighter_id  TEXT NOT NULL REFERENCES fighters(id) ON DELETE CASCADE,
  kata_name   TEXT NOT NULL DEFAULT '',
  panel_size  INTEGER NOT NULL DEFAULT 5,
  status      TEXT NOT NULL DEFAULT 'pending',
  total       REAL NOT NULL DEFAULT 0,
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE IF NOT EXISTS kata_scores (
  id             INTEGER PRIMARY KEY AUTOINCREMENT,
  performance_id TEXT NOT NULL REFERENCES kata_performances(id) ON DELETE CASCADE,
  judge          INTEGER NOT NULL,
  score          REAL NOT NULL,
  created_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  UNIQUE (performance_id, judge)
);

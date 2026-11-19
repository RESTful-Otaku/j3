-- Persistent bracket stages and slots with per-round history.

CREATE TABLE IF NOT EXISTS bracket_stages (
  id        TEXT PRIMARY KEY,
  event_id  TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  category_id TEXT REFERENCES categories(id) ON DELETE CASCADE,
  name      TEXT NOT NULL,
  kind      TEXT NOT NULL DEFAULT 'pool',
  sort_order INTEGER NOT NULL DEFAULT 0,
  status    TEXT NOT NULL DEFAULT 'pending',
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE IF NOT EXISTS bracket_slots (
  id          TEXT PRIMARY KEY,
  stage_id    TEXT NOT NULL REFERENCES bracket_stages(id) ON DELETE CASCADE,
  position    INTEGER NOT NULL,
  fighter_id  TEXT REFERENCES fighters(id) ON DELETE SET NULL,
  state       TEXT NOT NULL DEFAULT 'pending',
  result_icon TEXT NOT NULL DEFAULT '',
  source_slot TEXT NOT NULL DEFAULT '',
  UNIQUE (stage_id, position)
);

-- Medal awards, including dual bronze.

CREATE TABLE IF NOT EXISTS medals (
  id          TEXT PRIMARY KEY,
  event_id    TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  category_id TEXT,
  fighter_id  TEXT NOT NULL REFERENCES fighters(id) ON DELETE CASCADE,
  kind        TEXT NOT NULL CHECK (kind IN ('gold','silver','bronze','champion')),
  awarded_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

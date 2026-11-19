-- Event timeline with drift-aware projections.

CREATE TABLE IF NOT EXISTS schedule_items (
  id              TEXT PRIMARY KEY,
  event_id        TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  name            TEXT NOT NULL,
  area            TEXT NOT NULL DEFAULT '',
  starts_at       TEXT NOT NULL,
  duration_seconds INTEGER NOT NULL DEFAULT 300,
  status          TEXT NOT NULL DEFAULT 'upcoming',
  sort_order      INTEGER NOT NULL DEFAULT 0,
  actual_start    TEXT,
  actual_end      TEXT
);

CREATE TABLE IF NOT EXISTS schedule_state (
  event_id       TEXT PRIMARY KEY REFERENCES events(id) ON DELETE CASCADE,
  offset_minutes INTEGER NOT NULL DEFAULT 0
);

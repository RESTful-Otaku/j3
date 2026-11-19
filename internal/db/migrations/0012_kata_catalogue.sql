-- Kata catalogue and on-the-day training / sparring sessions.

CREATE TABLE IF NOT EXISTS kata_catalogue (
  id         TEXT PRIMARY KEY,
  name       TEXT NOT NULL,
  style      TEXT NOT NULL DEFAULT 'JJJ',
  level      TEXT NOT NULL DEFAULT 'standard',
  discipline TEXT NOT NULL DEFAULT 'kata',
  is_weapon  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS training_sessions (
  id              TEXT PRIMARY KEY,
  event_id        TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  title           TEXT NOT NULL,
  kind            TEXT NOT NULL DEFAULT 'drill',
  leader          TEXT NOT NULL DEFAULT '',
  starts_at       TEXT NOT NULL DEFAULT '',
  duration_seconds INTEGER NOT NULL DEFAULT 1800,
  status          TEXT NOT NULL DEFAULT 'scheduled'
);

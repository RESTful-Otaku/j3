-- Server-authoritative fight state and scoring events.

CREATE TABLE IF NOT EXISTS fight_state (
  fight_id          TEXT PRIMARY KEY REFERENCES fights(id) ON DELETE CASCADE,
  seconds_remaining INTEGER NOT NULL DEFAULT 180,
  running           INTEGER NOT NULL DEFAULT 0,
  round             INTEGER NOT NULL DEFAULT 1,
  updated_at        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE IF NOT EXISTS score_events (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  fight_id      TEXT NOT NULL REFERENCES fights(id) ON DELETE CASCADE,
  seq           INTEGER NOT NULL,
  kind          TEXT NOT NULL,
  corner        TEXT NOT NULL DEFAULT '',
  points        INTEGER NOT NULL DEFAULT 0,
  technique     TEXT NOT NULL DEFAULT '',
  card          TEXT NOT NULL DEFAULT '',
  actor_user_id TEXT NOT NULL DEFAULT 'u-official',
  payload_json  TEXT NOT NULL DEFAULT '{}',
  created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  UNIQUE (fight_id, seq)
);

CREATE INDEX IF NOT EXISTS score_events_fight ON score_events(fight_id, seq);

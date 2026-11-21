-- Full event configuration: costs, amenities, kit, multi-phase rules and mat staffing.

CREATE TABLE IF NOT EXISTS event_info (
  event_id       TEXT PRIMARY KEY REFERENCES events(id) ON DELETE CASCADE,
  organiser      TEXT NOT NULL DEFAULT '',
  venue          TEXT NOT NULL DEFAULT '',
  costs_json     TEXT NOT NULL DEFAULT '[]',
  refreshments   TEXT NOT NULL DEFAULT '',
  parking        TEXT NOT NULL DEFAULT '',
  hotel_json     TEXT NOT NULL DEFAULT '[]',
  evening_json   TEXT NOT NULL DEFAULT '{}',
  course_json    TEXT NOT NULL DEFAULT '{}',
  attendance_rule TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS kit_requirements (
  id          TEXT PRIMARY KEY,
  event_id    TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  item        TEXT NOT NULL,
  note        TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS mat_staff (
  id        TEXT PRIMARY KEY,
  event_id  TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  mat_name  TEXT NOT NULL,
  role      TEXT NOT NULL,
  user_id   TEXT REFERENCES users(id) ON DELETE SET NULL,
  sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS discipline_rules (
  id          TEXT PRIMARY KEY,
  event_id    TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  kind        TEXT NOT NULL DEFAULT 'fight',
  min_age     INTEGER NOT NULL DEFAULT 0,
  max_age     INTEGER NOT NULL DEFAULT 99,
  min_belt_order INTEGER NOT NULL DEFAULT 0,
  phases_json TEXT NOT NULL DEFAULT '[]',
  scoring_json TEXT NOT NULL DEFAULT '{}',
  kit_policy  TEXT NOT NULL DEFAULT '',
  notes       TEXT NOT NULL DEFAULT '',
  rules_json  TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS kata_sections (
  id           TEXT PRIMARY KEY,
  event_id     TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  name         TEXT NOT NULL,
  max_score    REAL NOT NULL DEFAULT 10,
  sort_order   INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS kata_section_scores (
  performance_id TEXT NOT NULL REFERENCES kata_performances(id) ON DELETE CASCADE,
  section_id     TEXT NOT NULL REFERENCES kata_sections(id) ON DELETE CASCADE,
  judge          INTEGER NOT NULL,
  score          REAL NOT NULL DEFAULT 0,
  PRIMARY KEY (performance_id, section_id, judge)
);

CREATE TABLE IF NOT EXISTS password_resets (
  id         TEXT PRIMARY KEY,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token      TEXT NOT NULL UNIQUE,
  expires_at TEXT NOT NULL,
  used       INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS course_attendance (
  entrant_id   TEXT PRIMARY KEY REFERENCES entrants(id) ON DELETE CASCADE,
  attended     INTEGER NOT NULL DEFAULT 0,
  excluded     INTEGER NOT NULL DEFAULT 0
);

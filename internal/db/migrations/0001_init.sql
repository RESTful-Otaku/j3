-- J3 initial schema: tenants, people, events, config snapshots, fights and audit.

CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  name    TEXT NOT NULL,
  applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE tenants (
  id         TEXT PRIMARY KEY,
  name       TEXT NOT NULL,
  slug       TEXT NOT NULL UNIQUE,
  locale     TEXT NOT NULL DEFAULT 'en',
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE themes (
  id          TEXT PRIMARY KEY,
  tenant_id   TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  logo_path   TEXT NOT NULL DEFAULT '',
  banner_path TEXT NOT NULL DEFAULT '',
  tokens_json TEXT NOT NULL DEFAULT '{}',
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE clubs (
  id            TEXT PRIMARY KEY,
  tenant_id     TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name          TEXT NOT NULL,
  logo_path     TEXT NOT NULL DEFAULT '',
  primary_colour TEXT NOT NULL DEFAULT '#2563eb',
  created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- RBAC roles are a fixed, typed hierarchy (see internal/auth/roles.go).
CREATE TABLE users (
  id            TEXT PRIMARY KEY,
  tenant_id     TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  club_id       TEXT REFERENCES clubs(id) ON DELETE SET NULL,
  role          TEXT NOT NULL CHECK (role IN
                ('tournament_runner','official','club_leader','fighter','spectator')),
  display_name  TEXT NOT NULL,
  email         TEXT NOT NULL,
  password_hash TEXT NOT NULL DEFAULT '',
  source        TEXT NOT NULL DEFAULT 'independent' CHECK (source IN ('independent','club_invite')),
  status        TEXT NOT NULL DEFAULT 'invited' CHECK (status IN ('invited','active','suspended')),
  created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX users_tenant_email ON users(tenant_id, lower(email));

-- Full JJJ belt spectrum is data, not an enum: built-in + custom ranks per tenant.
CREATE TABLE belt_ranks (
  id          TEXT PRIMARY KEY,
  tenant_id   TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  label       TEXT NOT NULL,
  colour      TEXT NOT NULL DEFAULT '#f8fafc',
  sort_order  INTEGER NOT NULL DEFAULT 0,
  grade_kind  TEXT NOT NULL DEFAULT 'kyu' CHECK (grade_kind IN ('kyu','dan','youth','custom')),
  stripe_count INTEGER NOT NULL DEFAULT 0,
  min_age     INTEGER,
  max_age     INTEGER,
  is_custom   INTEGER NOT NULL DEFAULT 0,
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE fighters (
  id          TEXT PRIMARY KEY,
  tenant_id   TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id     TEXT UNIQUE REFERENCES users(id) ON DELETE SET NULL,
  club_id     TEXT REFERENCES clubs(id) ON DELETE SET NULL,
  belt_rank_id TEXT REFERENCES belt_ranks(id) ON DELETE SET NULL,
  display_name TEXT NOT NULL,
  date_of_birth TEXT NOT NULL DEFAULT '',
  weight_kg   REAL,
  gender      TEXT NOT NULL DEFAULT 'open' CHECK (gender IN ('male','female','open','coed')),
  avatar_path TEXT NOT NULL DEFAULT '',
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- Versioned config snapshots keep historical events reproducible.
CREATE TABLE rule_sets (
  id          TEXT PRIMARY KEY,
  tenant_id   TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  version     INTEGER NOT NULL DEFAULT 1,
  config_json TEXT NOT NULL DEFAULT '{}',
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE format_policies (
  id          TEXT PRIMARY KEY,
  tenant_id   TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  version     INTEGER NOT NULL DEFAULT 1,
  config_json TEXT NOT NULL DEFAULT '{}',
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE events (
  id               TEXT PRIMARY KEY,
  tenant_id        TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  theme_id         TEXT REFERENCES themes(id) ON DELETE SET NULL,
  rule_set_id      TEXT REFERENCES rule_sets(id) ON DELETE SET NULL,
  format_policy_id TEXT REFERENCES format_policies(id) ON DELETE SET NULL,
  name             TEXT NOT NULL,
  starts_at        TEXT NOT NULL DEFAULT '',
  ends_at          TEXT NOT NULL DEFAULT '',
  status           TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published','live','complete')),
  created_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE categories (
  id          TEXT PRIMARY KEY,
  event_id    TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  kind        TEXT NOT NULL CHECK (kind IN ('age','weight','belt','gender','combined')),
  min_value   REAL,
  max_value   REAL,
  belt_rank_id TEXT REFERENCES belt_ranks(id) ON DELETE SET NULL,
  gender      TEXT NOT NULL DEFAULT 'open',
  sort_order  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE entrants (
  id           TEXT PRIMARY KEY,
  event_id     TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  fighter_id   TEXT NOT NULL REFERENCES fighters(id) ON DELETE CASCADE,
  category_id  TEXT REFERENCES categories(id) ON DELETE SET NULL,
  club_id      TEXT REFERENCES clubs(id) ON DELETE SET NULL,
  seed         INTEGER,
  checked_in_at TEXT,
  weighed_in_kg REAL,
  created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  UNIQUE(event_id, fighter_id, category_id)
);

CREATE TABLE mats (
  id        TEXT PRIMARY KEY,
  event_id  TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  name      TEXT NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  status    TEXT NOT NULL DEFAULT 'idle' CHECK (status IN ('idle','live','paused','closed'))
);

CREATE TABLE fights (
  id              TEXT PRIMARY KEY,
  event_id        TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  mat_id          TEXT REFERENCES mats(id) ON DELETE SET NULL,
  category_id     TEXT REFERENCES categories(id) ON DELETE SET NULL,
  round_label     TEXT NOT NULL DEFAULT '',
  red_fighter_id  TEXT REFERENCES fighters(id) ON DELETE SET NULL,
  blue_fighter_id TEXT REFERENCES fighters(id) ON DELETE SET NULL,
  state           TEXT NOT NULL DEFAULT 'pending' CHECK (state IN ('pending','live','paused','complete','forfeit')),
  outcome         TEXT NOT NULL DEFAULT '' CHECK (outcome IN ('','win_red','win_blue','draw','dq_red','dq_blue')),
  best_of         INTEGER NOT NULL DEFAULT 1,
  version         INTEGER NOT NULL DEFAULT 1,
  created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  completed_at    TEXT
);

-- Live scoring events are committed as confirmed; displays query this table.
CREATE TABLE fight_events (
  fight_id     TEXT NOT NULL REFERENCES fights(id) ON DELETE CASCADE,
  seq          INTEGER NOT NULL,
  kind         TEXT NOT NULL,
  actor_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
  payload_json TEXT NOT NULL DEFAULT '{}',
  created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  PRIMARY KEY (fight_id, seq)
);

CREATE TABLE invitations (
  id         TEXT PRIMARY KEY,
  tenant_id  TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  event_id   TEXT REFERENCES events(id) ON DELETE CASCADE,
  club_id    TEXT REFERENCES clubs(id) ON DELETE SET NULL,
  role       TEXT NOT NULL,
  email      TEXT NOT NULL,
  token_hash TEXT NOT NULL,
  state      TEXT NOT NULL DEFAULT 'pending' CHECK (state IN ('pending','accepted','expired','revoked')),
  expires_at TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- accountability: every material action is recorded with actor, role and before/after.
CREATE TABLE audit_log (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id     TEXT NOT NULL,
  actor_user_id TEXT,
  actor_role    TEXT NOT NULL,
  action        TEXT NOT NULL,
  target_type   TEXT NOT NULL,
  target_id     TEXT NOT NULL,
  before_json   TEXT NOT NULL DEFAULT '',
  after_json    TEXT NOT NULL DEFAULT '',
  created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX audit_log_tenant_created ON audit_log(tenant_id, created_at);

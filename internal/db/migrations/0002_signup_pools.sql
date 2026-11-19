-- Self-service sign-up stats, role approvals and automated pool assignments.

ALTER TABLE users ADD COLUMN phone TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN date_of_birth TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN height_cm REAL;
ALTER TABLE users ADD COLUMN weight_kg REAL;
ALTER TABLE users ADD COLUMN height_unit TEXT NOT NULL DEFAULT 'metric';
ALTER TABLE users ADD COLUMN weight_unit TEXT NOT NULL DEFAULT 'metric';
ALTER TABLE users ADD COLUMN belt_rank_id TEXT REFERENCES belt_ranks(id);
ALTER TABLE users ADD COLUMN gender TEXT NOT NULL DEFAULT 'open';
ALTER TABLE users ADD COLUMN role_request TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN role_request_status TEXT NOT NULL DEFAULT 'none';

CREATE TABLE IF NOT EXISTS role_requests (
  id           TEXT PRIMARY KEY,
  tenant_id    TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  requested_role TEXT NOT NULL,
  status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected')),
  reviewed_by  TEXT REFERENCES users(id) ON DELETE SET NULL,
  reviewed_at  TEXT,
  created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE IF NOT EXISTS pools (
  id          TEXT PRIMARY KEY,
  event_id    TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  category_id TEXT REFERENCES categories(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  sort_order  INTEGER NOT NULL DEFAULT 0,
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE IF NOT EXISTS pool_members (
  pool_id     TEXT NOT NULL REFERENCES pools(id) ON DELETE CASCADE,
  fighter_id  TEXT NOT NULL REFERENCES fighters(id) ON DELETE CASCADE,
  seed        INTEGER,
  PRIMARY KEY (pool_id, fighter_id)
);

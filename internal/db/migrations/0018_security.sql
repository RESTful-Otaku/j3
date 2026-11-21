-- Sessions, guardian consent and stricter identity data.
CREATE TABLE IF NOT EXISTS sessions (
  token     TEXT PRIMARY KEY,
  user_id   TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role      TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

ALTER TABLE fighters ADD COLUMN guardian_name TEXT NOT NULL DEFAULT '';
ALTER TABLE fighters ADD COLUMN guardian_contact TEXT NOT NULL DEFAULT '';
ALTER TABLE fighters ADD COLUMN guardian_consent INTEGER NOT NULL DEFAULT 0;
ALTER TABLE fighters ADD COLUMN medical_notes TEXT NOT NULL DEFAULT '';

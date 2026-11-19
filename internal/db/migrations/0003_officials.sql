-- Officials assigned to contests, shown on big-screen scoreboards.

CREATE TABLE IF NOT EXISTS fight_officials (
  fight_id TEXT NOT NULL REFERENCES fights(id) ON DELETE CASCADE,
  user_id  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role     TEXT NOT NULL CHECK (role IN ('referee','judge','scorekeeper')),
  PRIMARY KEY (fight_id, user_id, role)
);

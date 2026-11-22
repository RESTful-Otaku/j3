-- Authority chain: admins are flagged runners so the original role CHECK stays
-- valid on existing databases, while the API exposes a true admin role.
ALTER TABLE users ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN invited_by TEXT;
ALTER TABLE users ADD COLUMN must_change_password INTEGER NOT NULL DEFAULT 0;

ALTER TABLE invitations ADD COLUMN invited_by TEXT;
ALTER TABLE invitations ADD COLUMN actor_role TEXT NOT NULL DEFAULT '';
ALTER TABLE invitations ADD COLUMN accepted_at TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS invitations_token ON invitations(token_hash);

-- Admin has full control; other inviters are scoped in application code by
-- internal/auth.CanInvite (club owners may only add fighters to their club).
INSERT OR REPLACE INTO role_permissions (role, resource, can_view, can_create, can_edit, can_delete) VALUES
  ('admin', 'participants', 1, 1, 1, 1),
  ('admin', 'fights', 1, 1, 1, 1),
  ('admin', 'settings', 1, 1, 1, 1),
  ('admin', 'users', 1, 1, 1, 1),
  ('admin', 'invites', 1, 1, 1, 1);

INSERT OR IGNORE INTO role_permissions (role, resource, can_view, can_create, can_edit, can_delete) VALUES
  ('tournament_runner', 'users', 1, 1, 1, 0),
  ('tournament_runner', 'invites', 1, 1, 1, 1),
  ('official', 'users', 1, 1, 0, 0),
  ('official', 'invites', 1, 1, 0, 0),
  ('club_leader', 'users', 1, 1, 0, 0),
  ('club_leader', 'invites', 1, 1, 0, 0);

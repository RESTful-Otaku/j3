-- Configurable RBAC permissions per role.
CREATE TABLE IF NOT EXISTS role_permissions (
  role      TEXT NOT NULL,
  resource  TEXT NOT NULL,
  can_view  INTEGER NOT NULL DEFAULT 0,
  can_create INTEGER NOT NULL DEFAULT 0,
  can_edit  INTEGER NOT NULL DEFAULT 0,
  can_delete INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (role, resource)
);

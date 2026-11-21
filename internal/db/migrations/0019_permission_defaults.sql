-- Safe RBAC defaults, seeded once. Fighters and spectators have no edit/delete rights.
INSERT OR IGNORE INTO role_permissions (role, resource, can_view, can_create, can_edit, can_delete) VALUES
  ('tournament_runner', 'participants', 1, 1, 1, 1),
  ('tournament_runner', 'fights', 1, 1, 1, 1),
  ('tournament_runner', 'settings', 1, 1, 1, 1),
  ('official', 'participants', 1, 0, 0, 0),
  ('official', 'fights', 1, 1, 1, 0),
  ('club_leader', 'participants', 1, 1, 1, 0),
  ('club_leader', 'fights', 1, 0, 0, 0),
  ('club_leader', 'settings', 1, 1, 1, 0),
  ('fighter', 'participants', 1, 0, 0, 0),
  ('fighter', 'fights', 1, 0, 0, 0),
  ('spectator', 'participants', 0, 0, 0, 0),
  ('spectator', 'fights', 1, 0, 0, 0);

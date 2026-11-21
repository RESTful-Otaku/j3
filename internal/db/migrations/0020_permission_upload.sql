-- Club leaders may upload their own club branding.
INSERT OR REPLACE INTO role_permissions (role, resource, can_view, can_create, can_edit, can_delete)
VALUES ('club_leader', 'settings', 1, 1, 1, 0);

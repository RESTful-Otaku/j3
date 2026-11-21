-- Officials may review configuration and the accountability trail; runners may manage it.
INSERT OR REPLACE INTO role_permissions (role, resource, can_view, can_create, can_edit, can_delete)
VALUES ('official', 'settings', 1, 0, 0, 0);

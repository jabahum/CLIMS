-- Preserve the original CLIMS role IDs and spellings.
INSERT INTO roles (id, name, created_at, updated_at) VALUES
(1,'admin',NULL,NULL),
(2,'moh','2023-10-07 03:56:56','2023-10-07 03:56:56'),
(3,'nms','2023-10-07 03:57:17','2023-10-07 03:57:17'),
(4,'jms','2023-10-07 03:57:25','2023-10-07 03:57:25'),
(5,'dfcp','2023-10-07 03:57:41','2023-10-07 03:57:41'),
(6,'store incharge','2023-10-07 03:58:05','2023-10-07 03:58:05'),
(7,'hotspot manager','2023-10-07 03:58:40','2023-10-07 03:58:40'),
(8,'delivery agent','2023-10-07 03:59:41','2023-10-07 03:59:41'),
(9,'other supplier','2023-10-07 05:40:00','2023-10-07 05:40:00'),
(10,'healthy facility','2023-11-28 03:38:26','2023-11-28 03:38:26'),
(11,'vht','2023-11-29 08:51:12','2023-11-29 08:51:12')
ON CONFLICT (id) DO NOTHING;
SELECT setval(pg_get_serial_sequence('roles', 'id'), GREATEST((SELECT MAX(id) FROM roles), 11));

-- Fail rather than silently merge accounts with conflicting normalized emails.
CREATE UNIQUE INDEX users_email_normalized_unique ON users (lower(btrim(email))) WHERE email IS NOT NULL;

CREATE TABLE permissions (
    code text PRIMARY KEY,
    description text NOT NULL
);
CREATE TABLE role_permissions (
    role_id bigint NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_code text NOT NULL REFERENCES permissions(code) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_code)
);
INSERT INTO permissions (code, description) VALUES
('users:read', 'List users and view user accounts'),
('users:create', 'Create user accounts with an assigned role'),
('users:update', 'Activate or disable user accounts'),
('roles:read', 'List roles and their permissions'),
('roles:assign', 'Assign a role to a user'),
('roles:manage', 'Change permissions for non-admin roles');
INSERT INTO role_permissions (role_id, permission_code)
SELECT 1, code FROM permissions;

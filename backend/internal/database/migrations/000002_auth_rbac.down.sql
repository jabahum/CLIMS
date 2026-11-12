DROP TABLE role_permissions;
DROP TABLE permissions;
DROP INDEX users_email_normalized_unique;
-- The legacy role catalog remains: removing it could invalidate imported users.

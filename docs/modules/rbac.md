# RBAC module

[Module architecture](README.md) · [OpenAPI contract](../api/openapi.json)

Source: `backend/internal/features/rbac/`.

| File            | Responsibility                                                       |
| --------------- | -------------------------------------------------------------------- |
| `model.go`      | Role IDs, names, and permission lists                                |
| `dto.go`        | Complete permission-set replacement request                          |
| `repository.go` | Role/permission catalog queries and transactional permission updates |
| `middleware.go` | `PermissionHolder`, principal-local key, and permission gate         |
| `handler.go`    | Catalog responses, permission updates, and error handling            |
| `routes.go`     | Role and permission routes                                           |

## Role catalog

Migration 2 preserves the original CLIMS IDs and names:

| ID  | Role             |
| --- | ---------------- |
| 1   | admin            |
| 2   | moh              |
| 3   | nms              |
| 4   | jms              |
| 5   | dfcp             |
| 6   | store incharge   |
| 7   | hotspot manager  |
| 8   | delivery agent   |
| 9   | other supplier   |
| 10  | healthy facility |
| 11  | vht              |

`healthy facility` is the source backup's spelling. The API manages permissions for these roles; it does not expose role creation, renaming, or deletion.

## Permission policy

| Permission     | Capability                                    |
| -------------- | --------------------------------------------- |
| `users:read`   | List and read account records                 |
| `users:create` | Create accounts, together with `roles:assign` |
| `users:update` | Activate or disable accounts                  |
| `roles:read`   | List roles and available permission codes     |
| `roles:assign` | Assign user roles, including admin            |
| `roles:manage` | Replace a non-admin role's permissions        |

Admin starts with all six permissions. Other roles initially have no management permissions; every active account can still read its own profile. Role names alone do not grant additional domain access. Inventory, orders, facilities, and organization-scoped policies are not implemented yet.

`roles:assign` and `roles:manage` are delegation powers: a holder can grant broad access, including administrator-level access. They are intended for trusted account administrators.

## Routes and middleware

| Method | Route                            | Required permission |
| ------ | -------------------------------- | ------------------- |
| GET    | `/api/v1/roles`                  | `roles:read`        |
| GET    | `/api/v1/permissions`            | `roles:read`        |
| PUT    | `/api/v1/roles/{id}/permissions` | `roles:manage`      |

All routes require a valid JWT and an active account. `RequirePermission` reads `PermissionHolder` from the current-user local and returns 403 unless the permission is present. RBAC does not import auth or users; users implements the interface.

Permission replacement takes a role-row lock, deletes the previous mappings, inserts the requested codes, and commits as one transaction. Invalid codes roll back the entire update with 422. Duplicates are ignored. An empty array revokes all management permissions. Missing or null `permissions` is rejected, and requests are limited to 100 entries.

Role 1's permissions are immutable through this API; attempts to change them return 409. The users module separately protects the last active administrator from demotion or disabling. Protected requests reload account permissions from PostgreSQL, so a role or permission change affects the next request even with an existing JWT.

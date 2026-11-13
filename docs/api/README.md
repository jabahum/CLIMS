# API documentation

[Documentation index](../README.md)

## Authentication and RBAC API

All routes use `/api/v1`. Login and registration are public; all other endpoints require `Authorization: Bearer <token>`.

| Method | Path                       | Access                                                  |
| ------ | -------------------------- | ------------------------------------------------------- |
| POST   | `/auth/register`           | Public, creates an active account with no assigned role |
| POST   | `/auth/login`              | Public, returns a 24-hour JWT and account profile       |
| GET    | `/auth/me`                 | Any active account, including unassigned accounts       |
| GET    | `/users?limit=50&offset=0` | `users:read` (limit 1–100)                              |
| GET    | `/users/:id`               | `users:read`                                            |
| POST   | `/users`                   | Both `users:create` and `roles:assign`                  |
| PATCH  | `/users/:id/role`          | `roles:assign`                                          |
| PATCH  | `/users/:id/status`        | `users:update`                                          |
| GET    | `/roles`                   | `roles:read`                                            |
| GET    | `/permissions`             | `roles:read`                                            |
| PUT    | `/roles/:id/permissions`   | `roles:manage`                                          |

Registration accepts `username`, `email`, `password`, and optional `firstname`, `lastname`, and `phone_number`. Passwords require at least 8 characters and at most 72 UTF-8 bytes. Unknown JSON fields are rejected, including attempts to supply a role or status during public registration. Email uniqueness is case-insensitive. Existing nonunique usernames are preserved.

```sh
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"newuser","email":"user@example.com","password":"choose-a-strong-password"}'

curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"choose-a-strong-password"}'
```

Admin-created accounts use the registration fields plus `role_id`. Assign a role with `{"role_id":7}` and change account status with `{"status":"active"}` or `{"status":"disabled"}`. Replace a role's permissions with `{"permissions":["users:read"]}`; an empty array revokes all its management permissions. Unknown permission codes are rejected atomically.

The catalog preserves the 11 original roles: `admin` (1), `moh` (2), `nms` (3), `jms` (4), `dfcp` (5), `store incharge` (6), `hotspot manager` (7), `delivery agent` (8), `other supplier` (9), `healthy facility` (10), and `vht` (11). The legacy spelling `healthy facility` is intentional.

Admin initially has all six management permissions. Other roles initially have access to their own profile; an admin can grant them management permissions. Inventory, ordering, and organization-scoped access will need policies alongside their feature APIs. Permissions such as `roles:assign` and `roles:manage` delegate authority to grant access, including admin-level access; assign them only to trusted account administrators.

Permissions and account status are read from PostgreSQL for each protected request. Changes apply to existing JWTs on the next request. Disabled accounts cannot log in or use an existing JWT. The API prevents disabling or demoting the last active admin, including concurrent requests, and keeps the admin role's permission set fixed. There are no refresh tokens or logout/session revocation endpoints yet; disabling an account blocks its tokens while it remains disabled.

Login allows 10 failed attempts per IP per minute; registration allows 10 requests per IP per minute. These limits are per backend process and reset on restart. Deployments with multiple instances should use shared limiter storage. Forwarded headers are not trusted as client identity by default.

### Provision an administrator

Existing imported active users with legacy `role='1'` retain admin access. On a fresh installation, use the operator command with access to the database. Set `CLIMS_ADMIN_EMAIL` and `CLIMS_ADMIN_PASSWORD` in your shell (and optionally `CLIMS_ADMIN_USERNAME`, default `admin`), then run:

```sh
make create-admin
```

The command applies migrations and inserts a new administrator; it does not overwrite an existing account. Credentials are read from the environment and passwords are never printed. The command requires Go 1.26+ and uses the same `ENV_FILE` as the API.

### Run RBAC integration tests

`make test` runs unit tests. PostgreSQL integration tests skip unless `TEST_DATABASE_URL` is set. Point it to a **test server** whose user can create databases, then run:

```sh
make test-integration
```

The tests create and remove uniquely named databases on that server. They cover fresh installs, complete legacy data imports, upgrades from migration 1, legacy bcrypt passwords, registration privilege escalation, login, disabled accounts, permission assignment and revocation with existing tokens, invalid permission rollback, and concurrent protection of the last active admin.

## Interactive Swagger UI

Start the stack and visit <http://localhost:8080/docs/>. The UI loads the [OpenAPI 3.0.3 contract](openapi.json) from `/docs/openapi.json` and uses the API's current origin, including a custom host port. No external validator receives the specification. Swagger UI assets are served locally and require no CDN connection at runtime.

1. Call **Auth → Log in** with a valid account's email and password.
2. Copy `data.token` from the response.
3. Click **Authorize** and paste the token without adding a `Bearer ` prefix.
4. Call protected endpoints. Required application permissions appear in each operation's description.

Authorization is kept in page memory and is not persisted after a reload. “Try it out” sends real requests: registration creates accounts, and management operations can change access. Use a development database when exploring write operations.

Health operations are outside `/api/v1`: `/health/live` and `/health/ready` are public. Swagger UI and its specification are public documentation routes.

## Response and error formats

Successful API responses use `{"success":true,"message":"...","data":...}`. Mutations without response data omit `data`. Most errors use `{"success":false,"message":"..."}` with optional `error` details. Validation details map field names to messages.

JWT middleware also returns `{"error":"Unauthorized: missing or invalid token"}` for missing, invalid, or expired tokens. An invalid JWT subject can return plain-text `Unauthorized`. The OpenAPI 401 response documents all these formats. A disabled account or missing permission returns 403; a failed database lookup returns 500 rather than an authorization denial.

## Source, serving, and updates

`docs/api/openapi.json` is the authoritative, hand-maintained API contract. There is no generated duplicate in the backend. The `apidocs` feature serves an allowlist of UI and specification files; it does not expose the repository or module Markdown guides.

Local Go execution defaults `DOCS_DIR` to `../docs`, relative to the backend working directory. For a standalone binary, ship the root docs folder and set `DOCS_DIR` to its absolute path. Docker builds copy `docs/api` into the image and set `DOCS_DIR=/app/docs`. Documentation-file failures return 404 for docs assets and do not prevent the API from starting.

Swagger UI's exact version, upstream source, and SHA-256 hashes are recorded in [`assets/manifest.json`](assets/manifest.json); its license is included in [`assets/LICENSE`](assets/LICENSE). Update the bundle, stylesheet, license, and manifest together when upgrading.

When changing an endpoint, update its operation, request/response schemas, security requirement, status codes, and `x-required-permissions` metadata. Permission arrays mean **all** listed permissions are required; they are not OAuth scopes. Keep module guides in sync.

Run `make test` to check that every implemented auth, users, RBAC, and health route has an OpenAPI operation, that no documented operation lacks a route, and that documentation assets are served correctly. `make docs-check` additionally validates the OpenAPI document with a pinned validator in Docker and checks documentation links and asset checksums. No database is required for documentation checks.

For database-backed regression coverage, set `TEST_DATABASE_URL` to a disposable PostgreSQL server and run `make test-integration`.

References: [OpenAPI 3.0.3](https://spec.openapis.org/oas/v3.0.3.html) and [Swagger UI distribution](https://swagger.io/docs/open-source-tools/swagger-ui/usage/installation/).

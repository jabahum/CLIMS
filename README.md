# CLIMS

The runnable stack currently contains a Go/Fiber API and PostgreSQL 16. The `frontend/` and `mobile/` directories are placeholders. The default Compose stack runs the implemented services; the existing frontend Dockerfiles and Nginx configuration are reserved for later frontend work.

## Start locally

Requires Docker Desktop with Compose v2 (`include` support requires v2.20+). Run commands from the repository root:

```sh
make setup
make docker-up
```

`make setup` copies `infra/env/dev.env.example` only if your environment file does not exist. The backend embeds and applies versioned schema migrations at startup. PostgreSQL stores its data in a named volume; restarting the stack preserves it.

- API: `http://localhost:8080/api/v1`
- Liveness: `http://localhost:8080/health/live`
- Database readiness: `http://localhost:8080/health/ready`
- PostgreSQL: `localhost:5432`, database `clims_db`, user `clims_user`

Development credentials are in `infra/env/dev.env`. Change passwords and `JWT_SECRET` for deployment. Host ports bind to localhost. `DB_PORT` and `PORT` control host ports in Compose; containers always use 5432 and 8080 respectively. Compose sets the backend's database host to `postgres`; `DB_HOST=localhost` in the env file supports running Go directly.

```sh
make docker-down       # stops the stack, keeps database data
make docker-build      # rebuilds and starts services
make db-up             # PostgreSQL only
make run-backend       # requires Go 1.26+, loads the environment file
make test              # backend tests
```

Use `ENV_FILE=infra/env/prod.env` to select another environment file. Copy `infra/env/prod.env.example` and replace its credentials first. For an external PostgreSQL server, run the backend with `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`, and `JWT_SECRET` set. Set `DB_SSLMODE` for the server's TLS policy. The local Compose database uses `disable` on its private container network.

If Docker reports a permission error under `~/.docker/buildx/refs` (a root-owned buildx directory on this workstation), use a writable buildx configuration:

```sh
BUILDX_CONFIG=/tmp/clims-buildx make docker-up
```

## Import the legacy data

`infra/compose/clims_dump_11Nov2026.sql` is the original MySQL dump and must not be run against PostgreSQL. The converted PostgreSQL data is `infra/database/clims_postgres_seed.sql`. Schema and data are separate so normal startup creates an empty database without loading legacy accounts or tokens.

After the backend is healthy and has created the schema, import the legacy data **once into a database with no users or legacy records**, before registering any users (migration 2 prepopulates the role catalog):

```sh
make db-seed
```

The import has 29,898 rows across 18 tables. It runs in a transaction with `ON_ERROR_STOP`, preserves IDs and string escapes, and sets identity sequences beyond the imported IDs and original MySQL counters. It contains no DROP or TRUNCATE statements. The seed skips canonical role IDs already created by migration 2. A duplicate import fails on other existing keys and rolls back its row changes. Do not serve traffic while importing data.

The schema retains legacy column names and data types, including quantities/dates originally stored as strings, spelling such as `longtitude`, and existing password hashes in `users.password`. The Go auth repository uses this schema. No foreign keys have been invented for legacy varchar references. Emails are trimmed and compared case-insensitively by the auth API, with normalized uniqueness enforced by migration 2. Other PostgreSQL text comparisons remain case-sensitive; the original MySQL collations are not reproduced. Legacy usernames are not unique in the source data, so uniqueness is enforced on email only.

The source and converted dump contain account data and access tokens; keep them within the same access boundary as the database backup.

For a database outside Compose, apply the schema migration first (or start the API against that database), then import:

```sh
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f infra/database/clims_postgres_seed.sql
```

The conversion is reproducible:

```sh
make sql-convert
```

This regenerates migration 1, its rollback, the PostgreSQL seed, and `infra/database/seed_manifest.json` from the original dump. Do not regenerate already deployed migrations after changing the source dump; add a new migration instead. The converter targets the syntax in this specific backup and rejects unsupported table definitions and INSERT formats.

## Validate PostgreSQL compatibility

To verify without touching the application's database, start an isolated PostgreSQL container:

```sh
docker run -d --rm --name clims-postgres-validation \
  -e POSTGRES_PASSWORD=validation_only -e POSTGRES_DB=clims_validation \
  -v "$PWD:/workspace:ro" postgres:16-alpine
# Wait until this succeeds:
docker exec clims-postgres-validation pg_isready -U postgres -d clims_validation

docker exec clims-postgres-validation psql -U postgres -d clims_validation \
  -v ON_ERROR_STOP=1 -1 \
  -f /workspace/backend/internal/database/migrations/000001_init_schema.up.sql

docker exec clims-postgres-validation psql -U postgres -d clims_validation \
  -v ON_ERROR_STOP=1 -f /workspace/infra/database/clims_postgres_seed.sql

python3 scripts/verify_postgres.py clims-postgres-validation

docker stop clims-postgres-validation
```

The verifier compares every imported column value with the original dump and checks that generated IDs exceed the imported IDs. It consumes one sequence value per identity table; use it on a disposable validation database.

The PostgreSQL conversion uses [PostgreSQL 16 identity columns](https://www.postgresql.org/docs/16/sql-createtable.html). The JWT middleware uses the [official Fiber v3 module](https://docs.gofiber.io/contrib/jwt/).

To smoke-test authentication against a disposable running stack (creates a test account):

```sh
python3 scripts/smoke_api.py http://localhost:8080
```

## Backend modules

Feature modules live under `backend/internal/features/`:

- `auth`: public registration, login, JWT issuance, and authenticated profile routes.
- `users`: account models, persistence, shared account-creation service, active-account middleware, user administration, and role assignments.
- `rbac`: role and permission models, catalog persistence, role-permission administration, and permission middleware.

`auth` uses the `users` account service. `users` references RBAC roles and implements its permission-holder interface; `rbac` does not depend on `auth` or `users`. The server composes each module's routes with JWT and active-account middleware. Strict JSON decoding and positive-ID parsing are shared in `internal/pkg/request`. Cross-module PostgreSQL tests live in `internal/integration`. The admin CLI uses the same account-creation service as the API.

The module split preserves existing API paths, response shapes, database schema, rate limits, and permission rules.

## Authentication and RBAC API

All routes use `/api/v1`. Login and registration are public; all other endpoints require `Authorization: Bearer <token>`.

| Method | Path | Access |
| --- | --- | --- |
| POST | `/auth/register` | Public, creates an active account with no assigned role |
| POST | `/auth/login` | Public, returns a 24-hour JWT and account profile |
| GET | `/auth/me` | Any active account, including unassigned accounts |
| GET | `/users?limit=50&offset=0` | `users:read` (limit 1–100) |
| GET | `/users/:id` | `users:read` |
| POST | `/users` | Both `users:create` and `roles:assign` |
| PATCH | `/users/:id/role` | `roles:assign` |
| PATCH | `/users/:id/status` | `users:update` |
| GET | `/roles` | `roles:read` |
| GET | `/permissions` | `roles:read` |
| PUT | `/roles/:id/permissions` | `roles:manage` |

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

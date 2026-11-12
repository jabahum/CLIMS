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

After the backend is healthy and has created the schema, import the legacy data **once into an empty database**, before registering any users:

```sh
make db-seed
```

The import has 29,898 rows across 18 tables. It runs in a transaction with `ON_ERROR_STOP`, preserves IDs and string escapes, and sets identity sequences beyond the imported IDs and original MySQL counters. It contains no DROP or TRUNCATE statements. A duplicate import fails on existing keys and rolls back its row changes. Do not serve traffic while importing data.

The schema retains legacy column names and data types, including quantities/dates originally stored as strings, spelling such as `longtitude`, and existing password hashes in `users.password`. The Go auth repository uses this schema. No foreign keys have been invented for legacy varchar references. PostgreSQL uses its default case-sensitive text comparisons; the original MySQL case-insensitive collations are not reproduced. Legacy usernames are not unique in the source data, so uniqueness is enforced on email only.

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

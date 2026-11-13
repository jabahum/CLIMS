# Creating PostgreSQL migrations

[Documentation index](README.md) · [PostgreSQL and legacy imports](postgresql.md)

CLIMS uses [golang-migrate](https://github.com/golang-migrate/migrate/blob/v4.20.1/cmd/migrate/README.md) for SQL migration files and execution. `backend/go.mod` pins the library version (currently v4.20.1). The generator reads that pin and runs the matching CLI through `go run`; it needs no separately installed `migrate` binary and does not modify Go dependencies.

## Generate the next migration

Use Go 1.26+ and Python 3, and run commands from the repository root. The first CLI invocation may download Go modules and requires network access.

Stop the API before generating or editing migration SQL. The development Air watcher rebuilds on `.sql` changes, and API startup applies pending migrations. A running API could apply a new migration before you finish writing it. For the Compose stack, stop services while preparing the change:

```sh
make docker-down
make migrate-create NAME=add_orders_created_at_index
```

If running Go directly, stop that process too. `make docker-down` preserves database volumes. Use `ENVIRONMENT=prod` when stopping that selected environment; write and validate migrations in development before deploying them.

With the current migrations, this creates two empty files:

```text
backend/internal/database/migrations/000003_add_orders_created_at_index.up.sql
backend/internal/database/migrations/000003_add_orders_created_at_index.down.sql
```

The CLI chooses the next sequential number from existing files, padded to six digits. Names must use lowercase snake_case. Missing or invalid names fail before running the CLI. Generation itself only writes files; it neither reads database credentials nor connects to PostgreSQL. You can also run `python3 scripts/create_migration.py add_orders_created_at_index`.

Keep each up/down pair under the same version and name. Coordinate numbering across branches: if two branches generate version 3, rebase and renumber the **unapplied** pair before merging. Do not rename, reorder, or edit a migration already applied to a shared database.

## Write the SQL

The CLI does not infer schema changes. Fill both files with reviewed PostgreSQL SQL before starting the API or committing the pair. For example:

```sql
-- 000003_add_orders_created_at_index.up.sql
BEGIN;
CREATE INDEX orders_created_at_idx ON orders (created_at);
COMMIT;
```

```sql
-- 000003_add_orders_created_at_index.down.sql
BEGIN;
DROP INDEX orders_created_at_idx;
COMMIT;
```

This example indexes the existing orders table by creation time. Choose SQL appropriate to the actual schema change. Transactions make compatible changes atomic; PostgreSQL operations such as `CREATE INDEX CONCURRENTLY` require a different transaction strategy. See [PostgreSQL transaction restrictions for index creation](https://www.postgresql.org/docs/16/sql-createindex.html).

An up migration applies the change; its down migration should undo that change without undoing unrelated schema or data. Plan data transformations and rollbacks explicitly. Do not commit empty placeholders. Keep legacy data imports separate from normal schema migrations; `make sql-convert` regenerates the historical backup conversion and is not a migration generator.

## Validate against an isolated database

Use a disposable PostgreSQL 16 database with credentials permitted to create tables and indexes. Do not use the application's development/production database for rollback experiments. Stop any API process pointing at the test database so it cannot immediately reapply a rollback. From `backend/`, using the pinned CLI:

```sh
export TEST_DATABASE_URL='postgres://USER:PASSWORD@localhost:5432/clims_migration_test?sslmode=disable'

go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.20.1 \
  -path internal/database/migrations -database "$TEST_DATABASE_URL" up

go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.20.1 \
  -path internal/database/migrations -database "$TEST_DATABASE_URL" version

# Roll back only the newest migration, inspect the schema, then reapply it.
go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.20.1 \
  -path internal/database/migrations -database "$TEST_DATABASE_URL" down 1

go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.20.1 \
  -path internal/database/migrations -database "$TEST_DATABASE_URL" up 1
```

Update the explicit CLI version in these examples when upgrading the dependency. The existing integration suite also applies the current migration set on a fresh database and checks repeat application plus legacy seed compatibility:

```sh
# Repository root; this server must permit creating and dropping test databases.
TEST_DATABASE_URL='postgres://USER:PASSWORD@localhost:5432/postgres?sslmode=disable' \
  make test-integration
```

Add assertions for the schema/data behavior introduced by substantial migrations. Review the down path separately: the existing auth/RBAC tests do not automatically exercise every new rollback.

## Apply and deploy

After editing and validating the pair, restart development with `make dev-up`. The API embeds `backend/internal/database/migrations/*.sql` at compile time and calls `RunMigrations` on startup. It applies pending up migrations and accepts an already current database. A migration error stops startup/readiness.

Production images must be rebuilt/released to include new SQL; editing files beside an existing binary or container does not change embedded migrations. Migration sequence numbers are independent of the shared application `VERSION`. Schema state is recorded in PostgreSQL's `schema_migrations` table, including a dirty flag for failed applications.

If a migration fails, inspect the error, schema, and data before proceeding. The CLI's `force` command changes recorded migration state without running SQL; it does not repair a partial change. Use it only as part of a deliberate recovery plan after verifying the actual schema. Apply a new corrective migration for a change already deployed elsewhere, rather than rewriting migration history.

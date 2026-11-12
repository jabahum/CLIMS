# Development setup

[Documentation index](README.md)

## Start locally

Requires Docker Desktop with Compose v2 (`include` support requires v2.20+). Run commands from the repository root:

```sh
make setup
make docker-up
```

The Go toolchain must be 1.26+ when running or testing the backend outside Docker. Docker supplies its own toolchain.

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


See [PostgreSQL](postgresql.md), [modules](modules/README.md), and [API documentation](api/README.md) for the remaining setup and usage details.

Swagger UI is served by the API at `/docs/`. `DOCS_DIR` defaults to `../docs` for local Go execution and is set to `/app/docs` in the Docker image. See [API documentation](api/README.md) for standalone binary deployments.

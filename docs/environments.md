# Development and production environments

[Documentation index](README.md)

For published production images and version tags, see [service image releases](releases.md).

Run all commands from the repository root. `ENVIRONMENT=dev` is the default. Selecting `ENVIRONMENT=prod` switches both the Compose override and the default env file. `ENV_FILE` changes only the variables file.

## File layout

`infra/compose/docker-compose.yml` contains shared configuration. Always combine it with exactly one of `docker-compose.dev.yml` or `docker-compose.prod.yml` in that directory. The Makefile does this automatically. The old MariaDB and PostgreSQL Compose entrypoints have been retired; PostgreSQL is the database for both environments.

Every service has `Dockerfile.dev` and `Dockerfile.prod` under `infra/docker/<service>/`:

| Service  | Development                                                         | Production                                            |
| -------- | ------------------------------------------------------------------- | ----------------------------------------------------- |
| backend  | Go toolchain, Air reloads, mounted source/docs, module/build caches | Compiled binary and bundled API docs, nonroot runtime |
| frontend | Next.js dev server, mounted source, dependency/build volumes        | Standalone Next.js server, Node user, built assets    |
| mobile   | Expo Metro, mounted source, dependency volume                       | Optional iOS/Android JavaScript artifact export job   |
| postgres | PostgreSQL 16, localhost database port                              | PostgreSQL 16, internal database port only            |
| nginx    | Mounted development config, HTTP and Metro listeners                | Baked configuration, HTTP listener only               |

Both PostgreSQL Dockerfiles intentionally use the same upstream image. Runtime configuration belongs in Compose. Images are tagged `clims-<service>:dev` or `clims-<service>:prod`.

## Development

```sh
make setup
make dev-up
make check-stack
make dev-down
```

Development uses `infra/env/dev.env`, project name `clims`, and the existing `clims_postgres_data` volume. It publishes web/API/Swagger on localhost:8080, Metro on localhost:19000, and PostgreSQL on localhost:5432. Source changes reload Go, Next.js, and Metro. Nginx config edits require restarting Nginx. See [client setup](clients.md) for physical devices.

Client containers refresh dependencies from their lockfiles at startup. Cache volumes survive shutdown; database volumes also survive. Do not use `down -v` to refresh dependencies: it removes the database volume too.

## Production

```sh
make ENVIRONMENT=prod setup
# Edit infra/env/prod.env before starting services.
make ENVIRONMENT=prod compose-config
make prod-up
STACK_URL=http://localhost:8088 make ENVIRONMENT=prod check-stack
make prod-down
```

Set `DB_PASSWORD` and a strong `JWT_SECRET` in `infra/env/prod.env`; empty values fail Compose validation. The env file is ignored by Git. Set `NGINX_PORT` and `BIND_ADDRESS` for the deployment. The example binds localhost:8088, allowing both stacks to run on the same machine. To expose HTTP directly, configure the bind address explicitly. The supplied gateway serves HTTP; configure a TLS edge or certificates for an HTTPS deployment.

Production uses project name `clims-prod` and volume `clims-prod_postgres_data`, separate from development. It starts PostgreSQL, API, frontend, and Nginx. Server services have no source/config bind mounts, and only Nginx publishes a host port. Source and config changes require rebuilding the images. Existing database passwords are not changed by editing the env file: PostgreSQL initialization variables apply only to a new data directory.

`DB_SSLMODE=disable` applies to the bundled database on the private container network. For a separately hosted database, run the API with that server's connection and TLS settings; this Compose stack configures `DB_HOST=postgres` for its bundled service.

### Mobile release artifacts

Set `EXPO_PUBLIC_API_URL` in `infra/env/prod.env` to the deployed API address reachable by devices, then run:

```sh
make mobile-export
tar -tzf artifacts/mobile/mobile-bundles.tar.gz
```

The production mobile image requires a nonempty API URL at build time and exports both platforms to the ignored `artifacts/mobile/mobile-bundles.tar.gz`. Public Expo variables are embedded in the bundles; they must not contain secrets. The `artifacts` profile keeps this job out of normal production startup. This export produces JavaScript bundles and assets, not signed APK/IPA applications. Native release builds, signing, and distribution remain separate Expo/native workflows.

## Direct Compose and validation

Equivalent development command:

```sh
docker compose --env-file infra/env/dev.env \
  -f infra/compose/docker-compose.yml \
  -f infra/compose/docker-compose.dev.yml up --build -d --wait
```

Replace both `dev` occurrences with `prod` for production. Use the same files and project name for subsequent `logs`, `exec`, and `down` commands. A `-p` project override also changes the database volume namespace.

```sh
make compose-config
make check-environments
python3 scripts/check_docs.py
```

`check-environments` resolves both Compose models with disposable placeholder variables and checks Dockerfile parity, mounts, published ports, mobile profile/dependencies, and database isolation without starting containers or printing real secrets. `check-stack` checks the running HTTP routes; development also checks the Metro manifest. Set `STACK_URL` and, for development, `METRO_URL` when using custom ports.

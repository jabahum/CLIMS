# CLIMS documentation

The documentation source is this repository's root `docs/` folder. The root README is a navigation entry point.

See [CI checks and formatting](ci.md) for pull request/main checks and local verification commands.

See [service image releases](releases.md) for GHCR builds, version tags, and digest-pinned deployments.

See [creating PostgreSQL migrations](migrations.md) for pinned CLI scaffolding, SQL authoring, testing, and deployment.

| Guide                                    | Purpose                                                                          |
| ---------------------------------------- | -------------------------------------------------------------------------------- |
| [Setup](setup.md)                        | Start the stack, configure environments, and run the API locally                 |
| [Environments](environments.md)          | Development reloads, production images, Compose selection, and mobile exports    |
| [Web and mobile](clients.md)             | Next.js, Expo, Nginx routing, and device setup                                   |
| [PostgreSQL](postgresql.md)              | Migrations, legacy imports, and data verification                                |
| [Module architecture](modules/README.md) | Ownership, dependencies, request flow, and extension rules                       |
| [Auth](modules/auth.md)                  | Registration, login, JWTs, and rate limits                                       |
| [Users](modules/users.md)                | Accounts, creation service, administration, and account status                   |
| [RBAC](modules/rbac.md)                  | Legacy roles, permissions, and administrator protections                         |
| [API / Swagger](api/README.md)           | Endpoint reference, authentication examples, OpenAPI maintenance, and validation |

Run `make setup` and `make docker-up` from the repository root. Swagger UI is available at <http://localhost:8080/docs/> and its OpenAPI document at <http://localhost:8080/docs/openapi.json>.

The API contract is maintained in [`api/openapi.json`](api/openapi.json). It documents the implemented backend; it does not describe planned frontend, inventory, or organization-scoped features.

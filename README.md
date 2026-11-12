# CLIMS

CLIMS currently runs a Go/Fiber API with PostgreSQL 16. The frontend and mobile directories are placeholders.

Project documentation lives in the root [`docs/`](docs/README.md) folder:

- [Development setup](docs/setup.md)
- [PostgreSQL and legacy imports](docs/postgresql.md)
- [Module architecture](docs/modules/README.md)
- [Authentication module](docs/modules/auth.md)
- [Users module](docs/modules/users.md)
- [RBAC module](docs/modules/rbac.md)
- [API reference and Swagger](docs/api/README.md)

Run `make setup` and `make docker-up` from the repository root. Open [Swagger UI](http://localhost:8080/docs/) to explore the API.

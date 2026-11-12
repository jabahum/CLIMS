# CLIMS

CLIMS combines a Next.js web app, an Expo/React Native mobile project, a Go/Fiber API, and PostgreSQL 16. Docker Compose runs the stack behind Nginx, including the Metro development server.

Project documentation lives in the root [`docs/`](docs/README.md) folder:

- [Development setup](docs/setup.md)
- [PostgreSQL and legacy imports](docs/postgresql.md)
- [Module architecture](docs/modules/README.md)
- [Authentication module](docs/modules/auth.md)
- [Users module](docs/modules/users.md)
- [RBAC module](docs/modules/rbac.md)
- [Web and mobile projects](docs/clients.md)
- [API reference and Swagger](docs/api/README.md)

Run `make setup` and `make docker-up` from the repository root. Open [Swagger UI](http://localhost:8080/docs/) to explore the API.

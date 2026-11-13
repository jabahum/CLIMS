# Continuous integration

[Documentation index](README.md)

`.github/workflows/ci.yml` runs on every pull request and every push to `main`. It uses read-only repository permissions and requires no repository secrets. A newer run on the same ref cancels the older run.

| Check                           | Coverage                                                                                                                    |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| Formatting, docs, and Compose   | Prettier, Python Ruff lint/format, documentation links, vendored Swagger checksums, OpenAPI schema, and both Compose models |
| Go lint and PostgreSQL tests    | `gofmt`, `go vet`, race-enabled unit/integration tests against PostgreSQL 16, and command builds                            |
| frontend lint, types, and build | ESLint with zero warnings, TypeScript, and Next.js production build                                                         |
| mobile lint, types, and build   | ESLint with zero warnings, TypeScript, and Expo iOS/Android JavaScript export                                               |

The database job supplies `TEST_DATABASE_URL`, so integration tests run instead of skipping. Tests create and drop their own database, exercise migrations and legacy imports, and check authentication and RBAC. Its PostgreSQL service is disposable and uses CI-only credentials.

The clients currently have no component or end-to-end test suites; CI checks lint, types, and compilation/export. Mobile export does not build or sign APK/IPA files or run a simulator. CI validates Compose configuration without starting the complete Docker stack.

## Run checks locally

Use Go 1.26+, Node 22, Python 3.13, and Docker Compose v2. From the repository root:

```sh
npm ci
npm --prefix frontend ci
npm --prefix mobile ci
python3 -m venv .venv
. .venv/bin/activate
python -m pip install -r requirements-ci.txt

make format-check
make lint
make test
make check-frontend
make check-mobile
npm --prefix frontend run build
CI=1 EXPO_PUBLIC_API_URL=https://example.invalid/api/v1 npm --prefix mobile run export
python scripts/check_docs.py
python -m openapi_spec_validator docs/api/openapi.json
make check-environments
```

For the race-enabled database tests, use an isolated PostgreSQL server with permission to create databases:

```sh
cd backend
TEST_DATABASE_URL='postgres://USER:PASSWORD@localhost:5432/postgres?sslmode=disable' \
  go test -race -count=1 -timeout=5m ./...
```

Run `make format` to fix formatting. Prettier handles maintained JavaScript, TypeScript, CSS, JSON, YAML, HTML, and Markdown; Ruff formats Python; `gofmt` formats Go. The initial change normalizes existing files. Generated files, lockfiles, vendored Swagger assets, local editor settings, and generated agent guides are excluded from Prettier. Go and SQL are outside its supported parsers.

The root npm package contains tooling only; frontend and mobile retain independent lockfiles and installs. Keep formatter versions pinned in `package.json` and `requirements-ci.txt` so local and CI formatting agree. Refer to the [Prettier installation guide](https://prettier.io/docs/install) and [Ruff configuration guide](https://docs.astral.sh/ruff/configuration/) when changing these rules.

After this workflow is pushed and has run, its check names can be selected in GitHub branch protection if required. Adding the workflow does not change repository protection rules or deploy services.

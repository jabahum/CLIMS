# CLIMS documentation

The documentation source is this repository's root `docs/` folder. The root README is a navigation entry point.

See [CI checks and formatting](ci.md) for pull request/main checks and local verification commands.

See [service image releases](releases.md) for GHCR builds, version tags, and digest-pinned deployments.

See [creating PostgreSQL migrations](migrations.md) for pinned CLI scaffolding, SQL authoring, testing, and deployment.

| Guide                                             | Purpose                                                                                                    |
| ------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| [Setup](setup.md)                                 | Start the stack, configure environments, and run the API locally                                           |
| [Environments](environments.md)                   | Development reloads, production images, Compose selection, and mobile exports                              |
| [Web and mobile](clients.md)                      | Next.js, Expo, Nginx routing, and device setup                                                             |
| [PostgreSQL](postgresql.md)                       | Migrations, legacy imports, and data verification                                                          |
| [Module architecture](modules/README.md)          | Ownership, dependencies, request flow, and extension rules                                                 |
| [Auth](modules/auth.md)                           | Registration, login, JWTs, and rate limits                                                                 |
| [SSO and identity sync](sso-and-identity-sync.md) | Keycloak portal sign-in and MoH SSO Dashboard identity sync: roles, credentials, setup, troubleshooting    |
| [SSO troubleshooting](sso-troubleshooting.md)     | Diagnose `sso_error` codes in production, create the `clims` Keycloak client, recover after it was deleted |
| [Users](modules/users.md)                         | Accounts, creation service, administration, and account status                                             |
| [RBAC](modules/rbac.md)                           | Legacy roles, permissions, and administrator protections                                                   |
| [Hotspots](modules/hotspots.md)                   | Hotspot submission, review, activation and deactivation                                                    |
| [Locations](modules/locations.md)                 | Region, district and subcounty lookups shared across the system                                            |
| [Notifications](modules/notifications.md)         | Notifying people or groups from any feature, and the inbox API                                             |
| [Audit log](modules/audit.md)                     | Who did what, when and from where; recording and searching entries                                         |
| [Inventory](modules/inventory.md)                 | Stock reports for the warehouse, districts and hotspots, and recording stock                               |
| [Reports](modules/reports.md)                     | Programme reports: hotspot notification fulfilment, response times and the records behind them             |
| [API / Swagger](api/README.md)                    | Endpoint reference, authentication examples, OpenAPI maintenance, and validation                           |

Run `make setup` and `make docker-up` from the repository root. Swagger UI is available at <http://localhost:8080/docs/> and its OpenAPI document at <http://localhost:8080/docs/openapi.json>.

The API contract is maintained in [`api/openapi.json`](api/openapi.json). It documents the implemented backend; it does not describe planned frontend, inventory, or organization-scoped features.

## Documentation site

These files are also published as a Docusaurus site from [`website/`](../website/docusaurus.config.ts). With the development stack running it is at <http://localhost:8080/handbook/> and reloads as you edit. Without Docker, run `make docs-serve` (<http://localhost:3000/handbook/>); `make docs-build` builds it, and the build fails on broken links and anchors.

- Write plain Markdown (`.md`). It is parsed as CommonMark, so `<`, `>` and `{` need no escaping. Only `.mdx` files are parsed as MDX.
- Keep links relative to the file, as on GitHub. Links to files outside `docs/` are rewritten to GitHub when the site is built, and links to `api/openapi.json` go to the copy the site publishes. The site's [API reference](api/reference.mdx) page shows this folder's Swagger UI, with the same settings (`api/swagger-config.js`) as the API's `/docs/` page.
- New files in `modules/`, `adr/` and `releases/` appear in the sidebar automatically. Add a new top-level guide to [`website/sidebars.ts`](../website/sidebars.ts), or it is published without a sidebar entry.
- `website/package.json` pins every dependency exactly. Its `overrides` force patched versions of `tinypool`, `serialize-javascript` and `uuid`, which Docusaurus 3.10.2 uses only to build and serve the site, so `npm audit` is clean. When upgrading Docusaurus, drop each override its new release already covers.

# Auth module

[Module architecture](README.md) · [OpenAPI contract](../api/openapi.json)

Source: `backend/internal/features/auth/`.

| File              | Responsibility                                                                     |
| ----------------- | ---------------------------------------------------------------------------------- |
| `dto.go`          | Signup and login request shapes; signup embeds `users.AccountInput`                |
| `handler.go`      | Registration validation, credential checking, token issuance, own-profile response |
| `routes.go`       | Auth routes and per-process rate limiting                                          |
| `handler_test.go` | Login-attempt rate-limit regression test                                           |

## Endpoints

| Method | Route                   | Access                                       |
| ------ | ----------------------- | -------------------------------------------- |
| POST   | `/api/v1/auth/register` | Public; active, unassigned account           |
| POST   | `/api/v1/auth/login`    | Public; valid credentials and active account |
| GET    | `/api/v1/auth/me`       | Valid JWT and active account                 |

Registration trims usernames, trims and lowercases emails, rejects unknown fields, validates request fields, and delegates hashing and persistence to the users service. Clients cannot provide `role`, `role_id`, or `status`. The response contains `data.user_id`; registration does not issue a token.

Passwords need at least eight characters and at most 72 UTF-8 bytes. Validation's 72-character limit is supplemented by an explicit byte-length check. Existing legacy usernames can repeat; email uniqueness is enforced by PostgreSQL after trimming and lowercasing.

## Login and identity

Login looks up the account by normalized email and compares the password using bcrypt, including legacy `$2y$` hashes. Wrong credentials return 401. A correct password for a disabled account returns 403. Successful login returns a Bearer token, `expires_in: 86400`, and the current user profile, with the password hash excluded.

JWTs use HS256 with `JWT_SECRET`, contain `sub`, `iat`, and `exp`, and expire after 24 hours. `sub` is the user ID encoded as a string. Permissions are not embedded in the token; protected requests resolve them from PostgreSQL. The profile endpoint uses the current user already loaded by middleware.

Registration allows ten requests per IP per minute. Login allows ten failed requests per IP per minute; successful responses do not consume the limit. Limits are independent, in memory, per process, and reset on restart. Forwarded headers are not trusted as client identity by default.

## Boundaries and limitations

Auth does not own a repository or create role assignments. It uses the users repository for credential lookup and the shared users service for account creation. The server supplies JWT and active-account middleware during route registration.

There are no email-verification, password-reset, refresh-token, or logout/session-revocation APIs. Disabling an account blocks existing tokens while it remains disabled; reactivation allows an otherwise valid token again. See [users](users.md) and [RBAC](rbac.md) for administration.

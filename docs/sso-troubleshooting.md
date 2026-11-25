# SSO troubleshooting and client recovery

Runbook for portal sign-in (MoH Keycloak → CLIMS) when it breaks in production,
and for setting up or rebuilding the `clims` Keycloak client. Background,
credentials and file layout are in
[SSO and identity sync onboarding](sso-and-identity-sync.md); the sign-in flow
and identity-sync endpoints are in [infra/sso/README.md](../infra/sso/README.md).

Never print, paste or commit a client secret while following these steps. Every
check below compares secrets by fingerprint or by the status code Keycloak
returns.

## 1. Read the error code first

A failed portal sign-in lands on `/sign-in?sso_error=<code>`. The code says
which step of `GET /api/v1/auth/sso/callback`
(`backend/internal/features/auth/sso.go`) refused, so start there.

| Code          | Message on the sign-in page                                  | Step that refused                                                                 | Go to                         |
| ------------- | ------------------------------------------------------------ | --------------------------------------------------------------------------------- | ----------------------------- |
| `failed`      | Portal sign-in is not available right now                    | Exchanging the code with Keycloak, or a database error                            | [§3](#3-sso_errorfailed)      |
| `denied`      | Your portal account does not have access to CLIMS            | The ID token has no `clims_access` in `clims_roles`, or `azp` is wrong            | [§4](#4-sso_errordenied)      |
| `unavailable` | Your portal account is not linked to an active CLIMS account | No account link for this Keycloak user, or the account is inactive or has no role | [§5](#5-sso_errorunavailable) |
| `invalid`     | Portal sign-in could not be verified                         | Missing or mismatched state, nonce or ID token                                    | [§6](#6-other-symptoms)       |
| `expired`     | Portal sign-in took too long or was already used             | The 60-second handoff cookie was used or expired                                  | [§6](#6-other-symptoms)       |

The codes follow the order of the checks, so fixing one often reveals the next.
For example, after a wrong secret is fixed, `failed` becomes `denied` if the role
mapper is missing, then `unavailable` if the account isn't linked. Each new code
is progress.

## 2. Checks you can run without seeing a secret

Production paths: the env file is `<deploy root>/shared/prod.env`, the running
release is `<deploy root>/current`, and the backend container is
`clims-prod-backend-1` (see `scripts/deploy_release.sh`).

### Is the client secret right, and can the server reach Keycloak?

```sh
docker exec clims-prod-backend-1 sh -c \
  'curl -sS -o /dev/null -w "%{http_code}\n" "$OIDC_ISSUER/protocol/openid-connect/token" \
   -d grant_type=client_credentials -d client_id="$OIDC_CLIENT_ID" \
   --data-urlencode client_secret="$OIDC_CLIENT_SECRET"'
```

| Result              | Meaning                                                                                                                   |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `400`               | Secret accepted. The `clims` client has service accounts off, so Keycloak returns `unauthorized_client`; that is expected |
| `200`               | Secret accepted (service accounts are on for the client)                                                                  |
| `401`               | Wrong secret, or no client with that ID in the realm `OIDC_ISSUER` names                                                  |
| `000` or curl error | The server cannot reach Keycloak: DNS, firewall, proxy or TLS                                                             |

Drop `-o /dev/null -w ...` to see Keycloak's error body (for example
`invalid_client`); the body never contains the secret.

### Did the container pick up the secret in prod.env?

```sh
docker exec clims-prod-backend-1 sh -c 'printf %s "$OIDC_CLIENT_SECRET" | sha256sum | cut -c1-12'
grep '^OIDC_CLIENT_SECRET=' <deploy root>/shared/prod.env | cut -d= -f2- | tr -d '\n' | sha256sum | cut -c1-12
```

The two fingerprints must match. If they don't, the container was restarted
rather than recreated (see [§8](#8-apply-a-new-secret-to-production)).

### What does Keycloak put in the ID token?

Keycloak admin console → realm **MoH** → Clients → `clims` → Client scopes →
**Evaluate** → choose the user → **Generated ID token**. A token CLIMS accepts
contains:

```json
"azp": "clims",
"clims_roles": ["clims_access"]
```

### What did Keycloak see?

Realm settings → **Events** → turn on user events, try the sign-in again, then
open **Events**. A `CODE_TO_TOKEN_ERROR` with `invalid_client_credentials` is a
wrong secret; no event at all means the CLIMS server never reached Keycloak.

## 3. `sso_error=failed`

The browser came back from Keycloak with a code, but the backend could not turn
it into tokens, or the database failed while finding the account.

1. Run the [secret check](#is-the-client-secret-right-and-can-the-server-reach-keycloak).
2. `401`: the secret in prod differs from the one on the `clims` client. This
   always happens after the client is deleted and recreated, or after someone
   clicks **Regenerate** on its Credentials tab. Copy the secret from Keycloak
   and [apply it](#8-apply-a-new-secret-to-production).
3. `000`: fix network access from the server to `OIDC_ISSUER`.
4. `400` or `200` but still `failed`: the secret is fine, so the database
   lookup or session creation failed. Read
   `docker logs --since 5m clims-prod-backend-1` around the attempt.

## 4. `sso_error=denied`

The token exchange worked, but the ID token does not grant CLIMS access, the
token was issued to another client, or the user's access was revoked through
identity sync after the token was issued.

1. Check the token with [Evaluate](#what-does-keycloak-put-in-the-id-token).
2. No `clims_roles` claim at all: the mapper is missing. Clients → `clims` →
   Client scopes → `clims-dedicated` → Add mapper → By configuration →
   **User Client Role**:

   | Setting            | Value                |
   | ------------------ | -------------------- |
   | Name               | `CLIMS access roles` |
   | Client ID          | `clims`              |
   | Client Role prefix | empty                |
   | Multivalued        | On                   |
   | Token Claim Name   | `clims_roles`        |
   | Claim JSON Type    | String               |
   | Add to ID token    | On                   |

3. `clims_roles` present but without `clims_access`: the role doesn't exist or
   the user doesn't hold it. Clients → `clims` → Roles → `clims_access` must
   exist; Users (or the pilot group) → Role mapping → Assign role → filter by
   clients → `clims` `clims_access`.
4. `azp` is not `clims`: the user launched another client. Check the portal
   tile's launch URL is `https://<clims-host>/api/v1/auth/sso/start`.
5. Sign out of the portal so Keycloak issues a fresh token, then try again.

## 5. `sso_error=unavailable`

Keycloak and the token are fine. CLIMS has no link for this Keycloak issuer and
subject, or the linked CLIMS account is not active or has no role. Email is
never used to match accounts.

- Link the account through identity sync from the MoH SSO Dashboard, or record
  the link as described in [infra/sso/README.md](../infra/sso/README.md).
- In CLIMS, make sure the account is active and holds at least one role.

Deleting and recreating the `clims` client does **not** break account links:
links point at the Keycloak user (issuer and subject), not at the client.

## 6. Other symptoms

| Symptom                                                                 | Cause and fix                                                                                                                                                                                                                                                                                                                                           |
| ----------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `sso_error=invalid`                                                     | Stale tab, back button, or sign-in not started from the portal tile. Open CLIMS from the portal again                                                                                                                                                                                                                                                   |
| `sso_error=expired`                                                     | The handoff took longer than 60 seconds or was reused. Open CLIMS from the portal again                                                                                                                                                                                                                                                                 |
| Keycloak shows "Invalid redirect uri"                                   | The `clims` client's Valid redirect URIs must contain exactly `<OIDC_PUBLIC_ORIGIN>/api/v1/auth/sso/callback`                                                                                                                                                                                                                                           |
| CLIMS tile missing from the portal launcher                             | The user lacks `clims_access` (the client attribute `portal.accessRoles` gates the tile), or the client lost its `portal.*` and `ui.*` attributes. Common after the client was deleted; see [§7.2](#72-recover-after-the-clims-client-was-deleted)                                                                                                      |
| Keycloak admin console: "Failed to fetch dynamically imported module …" | The console page in the browser references assets from an older Keycloak build. Click "Press here to refresh and continue", hard-reload (Cmd/Ctrl+Shift+R) or use a private window. If it persists in a private window, a proxy in front of Keycloak is caching an old `index.html`; clear it, or use `kcadm.sh` ([§9](#9-the-same-steps-with-kcadmsh)) |
| CLIMS exits at startup about `OIDC_*`                                   | One of the four `OIDC_*` values is empty while `OIDC_CLIENT_ID` is set, `OIDC_PUBLIC_ORIGIN` has a path or trailing slash, or a URL uses HTTP outside localhost. Both URLs must be HTTPS in production                                                                                                                                                  |

## 7. Create or rebuild the `clims` client

Everything CLIMS needs from Keycloak, in the realm `OIDC_ISSUER` names:

| Item                                                                                                   | Where it lives                                         | Lost if the `clims` client is deleted     |
| ------------------------------------------------------------------------------------------------------ | ------------------------------------------------------ | ----------------------------------------- |
| `clims` client: confidential, standard flow, PKCE S256, redirect URI, `portal.*` and `ui.*` attributes | `infra/sso/clims.keycloak-client.json`                 | Yes                                       |
| `clims_roles` protocol mapper                                                                          | Same file                                              | Yes                                       |
| Client secret                                                                                          | Generated by Keycloak, copied into `prod.env`          | Yes: the new client gets a new secret     |
| `clims_access` and `identity_sync` client roles                                                        | `infra/sso/clims-identity-sync.keycloak-import.json`   | Yes                                       |
| Users' and groups' `clims_access` assignments                                                          | Keycloak role mappings                                 | Yes                                       |
| `identity_sync` on the sync client's service account and dedicated scope                               | Keycloak role and scope mappings                       | Yes                                       |
| `clims-identity-sync` client and its `clims` audience mapper                                           | Partial import file                                    | No (the mapper refers to `clims` by name) |
| Portal Systems entry                                                                                   | Dashboard registry, `infra/sso/clims.application.yaml` | No, but check it still launches           |
| CLIMS account links (`user_oidc_identities`)                                                           | CLIMS database                                         | No                                        |

### 7.1 Create the client in a new realm

A realm built from the dashboard's `infra/keycloak/realm-export.json` already
has all of this; only regenerate the two secrets. Otherwise:

1. Copy `infra/sso/clims.keycloak-client.json` and replace every
   `https://clims.example.org` with the CLIMS origin (for production,
   `https://clims.health.go.ug`):

   ```sh
   sed 's|https://clims.example.org|https://clims.health.go.ug|g' \
     infra/sso/clims.keycloak-client.json > /tmp/clims-client.json
   ```

2. Keycloak → realm → Clients → **Import client** → choose the file → Save.
3. Realm settings → Action → **Partial import** →
   `infra/sso/clims-identity-sync.keycloak-import.json`. This creates the
   `clims_access` and `identity_sync` roles on `clims` and the
   `clims-identity-sync` client. If `clims-identity-sync` already exists, choose
   **Skip** for existing resources.
4. Clients → `clims-identity-sync` → Service account roles → Assign role →
   `clims` → `identity_sync`.
5. Clients → `clims-identity-sync` → Client scopes →
   `clims-identity-sync-dedicated` → Scope → Assign role → `clims` →
   `identity_sync`. Without this every identity-sync call gets 403.
6. Assign `clims` → `clims_access` to the pilot users or group.
7. Clients → `clims` → Credentials: copy the secret and
   [apply it to production](#8-apply-a-new-secret-to-production). Do the same
   for `clims-identity-sync` on the dashboard side if its secret is new.
8. Portal → Admin → Systems: import `infra/sso/clims.application.yaml` (Import
   manifest) if CLIMS isn't registered, or open **Manage** on the CLIMS row and
   save it.
9. Verify: the [secret check](#is-the-client-secret-right-and-can-the-server-reach-keycloak)
   returns `400` or `200`, [Evaluate](#what-does-keycloak-put-in-the-id-token)
   shows `clims_roles: ["clims_access"]`, the tile appears after signing out of
   the portal and back in, and launching it signs in.

### 7.2 Recover after the `clims` client was deleted

Symptoms: the CLIMS tile disappears from the launcher, and portal sign-in fails
with `failed` (new secret), then `denied` (no mapper or role). Identity sync
also starts returning 403, because the sync client's `identity_sync` role was
deleted with `clims`.

Recreating the client by hand with **Create client** restores none of the items
marked "Yes" in the table above. Rebuild it from the files instead:

1. If a hand-made `clims` client exists, delete it (it holds nothing that the
   import doesn't bring back).
2. Follow [§7.1](#71-create-the-client-in-a-new-realm) steps 1 to 9. In step 3
   choose **Skip** so the existing `clims-identity-sync` client is kept, and in
   step 6 reassign `clims_access` to **everyone who had it**: the assignments
   are gone, so list the pilot group or users from the portal before deleting
   anything if you can.
3. The CLIMS account links survive, so once the steps are done users sign in to
   the same CLIMS accounts as before. No change is needed in CLIMS other than
   the new secret.

## 8. Apply a new secret to production

1. Edit `<deploy root>/shared/prod.env` in an editor (`nano`, `vi`) and replace
   the `OIDC_CLIENT_SECRET=` value. Avoid `echo`/`sed` with the secret on the
   command line, which leaves it in shell history.
2. Recreate the backend. `docker restart` keeps the old environment; Compose has
   to recreate the container:

   ```sh
   cd <deploy root>/current
   docker compose --project-name clims-prod \
     --env-file <deploy root>/shared/prod.env --env-file images.env \
     -f infra/compose/docker-compose.yml -f infra/compose/docker-compose.prod.yml \
     up -d --no-build --wait backend
   ```

3. Run the [fingerprint check](#did-the-container-pick-up-the-secret-in-prodenv)
   and the [secret check](#is-the-client-secret-right-and-can-the-server-reach-keycloak).
4. Update the local copy (`infra/env/prod.sso.env`, gitignored) so it doesn't
   reintroduce the old secret later.

## 9. The same steps with `kcadm.sh`

Use these when the admin console won't load. Run inside the Keycloak container;
`<clims client uuid>` is the ID in the console URL `…/clients/<uuid>/…`, or
from `kcadm.sh get clients -r MoH -q clientId=clims --fields id`.

```sh
kcadm.sh config credentials --server https://auth.health.go.ug --realm master --user <admin-user>

# Import the client (file from §7.1 step 1, copied into the container)
kcadm.sh create clients -r MoH -f /tmp/clims-client.json

CID=$(kcadm.sh get clients -r MoH -q clientId=clims --fields id --format csv --noquotes)

# Roles, if the partial import isn't used
kcadm.sh create clients/$CID/roles -r MoH -s name=clims_access
kcadm.sh create clients/$CID/roles -r MoH -s name=identity_sync

# Give a user CLIMS access
kcadm.sh add-roles -r MoH --uusername <keycloak-username> --cclientid clims --rolename clims_access

# Identity sync: service-account role and scope mapping
kcadm.sh add-roles -r MoH --uusername service-account-clims-identity-sync --cclientid clims --rolename identity_sync
SYNC=$(kcadm.sh get clients -r MoH -q clientId=clims-identity-sync --fields id --format csv --noquotes)
kcadm.sh create clients/$SYNC/scope-mappings/clients/$CID -r MoH \
  -b "[$(kcadm.sh get clients/$CID/roles/identity_sync -r MoH)]"

# Check the mapper is there
kcadm.sh get clients/$CID/protocol-mappers/models -r MoH --fields name,protocolMapper
```

`kcadm.sh` asks for the admin password interactively; don't pass it with
`--password` on the command line.

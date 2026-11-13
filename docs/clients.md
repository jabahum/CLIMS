# Web and mobile projects

[Documentation index](README.md)

The branch scaffolds clients and the gateway; it does not add login or user-management screens. Both clients include a CLIMS welcome screen and a real API readiness check.

## Projects and toolchains

| Folder      | Project                               | Current baseline                             |
| ----------- | ------------------------------------- | -------------------------------------------- |
| `frontend/` | Next.js App Router, TypeScript, React | Next.js 16.4.6, React 19.3                   |
| `mobile/`   | Expo / React Native, TypeScript       | Expo SDK 58, React Native 0.88.1, React 19.3 |

Use Node.js 22.13+ on the 22.x line (or another version supported by Expo SDK 58) for host development. Docker supplies Node.js 22. With nvm, run `nvm install` and `nvm use` from the repository root; `.nvmrc` selects Node 22. Both projects use npm and committed lockfiles. Keep application documentation in root `docs/`; the generated `AGENTS.md` files provide tool-specific project instructions.

## Run the whole stack

From the repository root:

```sh
make setup
make docker-up
make check-stack
```

For custom gateway ports, run the checker with matching `STACK_URL` and `METRO_URL`, for example `STACK_URL=http://localhost:8090 METRO_URL=http://localhost:19001 make check-stack`.

On this workstation, use `BUILDX_CONFIG=/tmp/clims-buildx make docker-up` if Docker's existing buildx metadata is root-owned.

| Service             | Container port | Host access                               |
| ------------------- | -------------- | ----------------------------------------- |
| Nginx HTTP gateway  | 80             | `http://localhost:8080`                   |
| Nginx Metro gateway | 8081           | `http://localhost:19000`                  |
| Next.js             | 3000           | Internal, through Nginx                   |
| Go API              | 8080           | Internal, through Nginx                   |
| Expo Metro          | 8081           | Internal, through Nginx                   |
| PostgreSQL          | 5432           | `localhost:5432` for local database tools |

Nginx routes `/` and Next.js assets to the frontend; `/api`, `/docs`, and `/health` retain their full paths and go to Go. A separate Nginx listener proxies Metro's root paths and WebSocket upgrades. Docker DNS resolution lets the proxy refresh upstream addresses after container recreation.

The startup sequence waits for PostgreSQL, then the API, then Next.js; Nginx waits for the API, frontend, and Metro health checks. The PostgreSQL named volume is preserved. The API is no longer published directly on host port 8080: Nginx owns that port.

Development runs Next.js with source reloads, Go with Air, and Expo Metro. Source folders are mounted; dependencies and build caches live in named volumes. Client startup runs `npm ci --prefer-offline` so dependencies follow the committed lockfiles after a restart. Rebuild the relevant image after changing Dockerfiles or toolchains.

Production runs a standalone Next.js build and compiled Go binary. Its mobile Dockerfile exports release JavaScript bundles as an optional job; it runs no Metro server. See [development and production environments](environments.md) for configuration, release artifacts, and commands.

## Web development

```sh
make install-frontend
make run-frontend
```

Open `http://localhost:3000`. Next.js rewrites `/api`, `/docs`, and `/health` to `BACKEND_URL`, defaulting to the Nginx gateway at `http://localhost:8080`. Copy `frontend/.env.example` to `frontend/.env.local` to change it. Browser API calls use relative paths; no private database credentials are bundled into the client.

The development Compose override uses `infra/docker/frontend/Dockerfile.dev`; production uses `Dockerfile.prod`. The web app uses system fonts and does not download Google Fonts.

```sh
make check-frontend
make build-frontend
```

## Mobile development and devices

```sh
make install-mobile
make run-mobile
```

The host command starts Expo directly (usually on port 8081). The Compose version is proxied at port 19000 to avoid collisions with other tools on 8081. Stop the Compose mobile service before running a second bundler against the same source if file watches or reloads conflict.

For the Compose server, open `exp://localhost:19000` in a compatible iOS simulator or use a physical-device configuration below. Use an Expo Go version supporting SDK 58, or create a matching development build. Android emulators and physical devices have their own loopback interfaces; `localhost` is not your computer's address on those devices.

For a physical device on the same LAN, set these in `infra/env/dev.env` (replace the example address with your computer's LAN IP):

```dotenv
BIND_ADDRESS=0.0.0.0
MOBILE_HOST=192.168.1.50
NGINX_PORT=8080
METRO_PORT=19000
EXPO_PUBLIC_API_URL=http://192.168.1.50:8080/api/v1
```

Then run `make docker-up` and open `exp://192.168.1.50:19000` in Expo Go. The computer and device must be on the same network, and the firewall must allow both gateway ports. `MOBILE_HOST` controls the advertised Metro URL; `EXPO_PUBLIC_API_URL` controls the API address bundled into the app. For Android emulators, a reachable LAN address also works; alternatively use the emulator's host mapping and configure both URLs accordingly.

When running Expo on the host, copy `mobile/.env.example` to `mobile/.env.local` and set `EXPO_PUBLIC_API_URL` to an API address reachable from the device. Public Expo variables are included in JavaScript bundles; use them for service URLs, never secrets. Restart Metro after changing environment variables. Custom gateway ports also require updating the API URL.

```sh
make check-mobile
cd mobile && npm run export
```

The export compiles JavaScript bundles for iOS and Android. It does not sign an app, create an APK/IPA, or verify rendering on a device. Native projects are generated by Expo tooling when needed; do not hand-create `ios/` or `android/` folders.

References: [Next.js installation](https://nextjs.org/docs/app/getting-started/installation), [Expo SDK 58](https://docs.expo.dev/versions/v58.0.0/), and [Expo proxy URL configuration](https://docs.expo.dev/more/expo-cli/#server-url).

## Dependency audit note

The frontend production dependency audit reports no advisories at setup time. Expo SDK 58 currently pulls `minimatch` 10 through `@expo/fingerprint`; npm reports the moderate [globparts denial-of-service advisory](https://github.com/advisories/GHSA-g3hv-75wg-m249) against that tooling dependency. The patched release is on minimatch 11, outside Expo's current major-version range. The scaffold preserves the SDK-supported dependency tree rather than forcing a cross-major override. Development lint dependencies also have moderate audit findings. Recheck these when updating the framework/tooling lockfiles.

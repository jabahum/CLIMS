# Service images and versioned releases

[Documentation index](README.md)

See the [2.0.0 release notes](releases/2.0.0.md) for the revamped platform release.

`.github/workflows/images.yml` builds the three application production Dockerfiles on pull requests without publishing. Pushing a version tag runs the complete [CI workflow](ci.md), publishes images to GitHub Container Registry (GHCR), and creates a GitHub Release after all three application images succeed. Normal pushes to `main` run CI; they do not publish images.

## Images and tags

Images use the repository's lowercase owner and name:

| Service  | Example release image                     |
| -------- | ----------------------------------------- |
| backend  | `ghcr.io/mohuganda/clims/backend:v2.0.0`  |
| frontend | `ghcr.io/mohuganda/clims/frontend:v2.0.0` |
| mobile   | `ghcr.io/mohuganda/clims/mobile:v2.0.0`   |

All three application images target `linux/amd64`, use production Dockerfiles, and include OCI source/version/commit labels, build provenance, and an SBOM. Builds cache each service independently. The backend receives version, commit, and timestamp build arguments. Development images are not published.

Supported tags are `vMAJOR.MINOR.PATCH`, optionally followed by prerelease identifiers, such as `v2.0.0-rc.1`. Numeric version components cannot have leading zeros. Build metadata (`+build`) is not supported because it cannot be used directly as a Docker tag. Prereleases are marked as such on GitHub. There are no floating `latest`, major, or minor image aliases.

The mobile image is an export job containing iOS and Android JavaScript bundles and assets. It does not run Metro or contain signed APK/IPA packages. Its API URL is embedded at build time; changing that URL requires a new image/version.

## Shared application version

The root `VERSION` file is authoritative, currently `2.0.0`. Backend, frontend, and mobile use one application version. Package manifests and lockfiles (including root tooling), the backend default version, Expo release metadata, and OpenAPI `info.version` must agree. CI checks this on every pull request and main push; the release preflight also requires the Git tag to be exactly `v` plus `VERSION`.

```sh
make version-set VERSION=2.1.0
make version-check
npm run format:check
```

Commit the version update with the release changes and merge it into `main` before creating `v2.1.0`. `make version-set VERSION=2.1.0-rc.1` supports prereleases too. Expo's native `version` keeps the numeric base (`2.1.0`), while `expo.extra.releaseVersion` and the mobile package version retain `2.1.0-rc.1`. This respects the [Expo/iOS version format](https://docs.expo.dev/versions/v58.0.0/config/app/#version). Native build numbers and signing remain separate workflows.

Image tags and OCI labels include the `v` prefix; application versions omit it. PR image tags use `pr-<number>` while the embedded application version remains the value in `VERSION`. Production Dockerfiles reject mismatching version build arguments; the backend defaults to the copied `VERSION` file and client images default to their checked package metadata. The Makefile reads `VERSION` for local builds too.

PostgreSQL and Nginx are infrastructure services with independent upstream versions. They are not built by the image workflow, published to GHCR, or included in release manifests. Their existing Dockerfiles and Compose services remain available. Build these infrastructure images locally before starting released application images, as shown below.

## Repository configuration

Set the repository Actions variable `EXPO_PUBLIC_API_URL` to the deployed API URL, for example `https://clims.example.org/api/v1`. Use a URL reachable by devices; public Expo variables must not contain secrets. Release validation rejects missing URLs, the `.invalid` test domain, embedded credentials, queries, and fragments. Pull request builds use `https://example.invalid/api/v1` instead and require no repository variable.

GHCR publication uses the built-in `GITHUB_TOKEN` with `packages: write`; release creation uses `contents: write` in its own job. No Docker Hub account or personal access token is required by the workflow. Existing GHCR packages must grant this repository Actions access if they were created elsewhere. Package visibility and who may pull images are controlled by GitHub package settings; see [GitHub's container publication guide](https://docs.github.com/en/actions/tutorials/publish-packages/publish-docker-images).

## Create a release

Merge the workflow and service changes into the publishing repository's `main` first. Then, from a clean checkout of that branch:

```sh
git switch main
git pull --ff-only origin main
git tag -a v2.0.0 -m 'Release v2.0.0'
git push origin v2.0.0
```

Use the next unused version; the example does not select a version automatically. A release tag must point to a commit reachable from `main`. The workflow rejects tags with an existing GitHub Release. Do not move published tags. Use a new version for changed code or build configuration.

After CI and all three application image uploads succeed, the workflow creates a GitHub Release with image references and two assets:

- `images.json`: version, application version, source commit, target platform, version tags, and image digests.
- `images.env`: `BACKEND_IMAGE`, `FRONTEND_IMAGE`, and `MOBILE_IMAGE`, each pinned by digest for Compose.

Image uploads are separate registry operations. If one fails, some images may already exist, but no GitHub Release is created. Fix transient failures and rerun failed jobs for the same unchanged tag. Rerunning an already published release is rejected. Consume the release's digest references to retain the exact image set even if registry tags are changed later.

## Run published images

Create `infra/env/prod.env` with deployment credentials as described in [environments](environments.md). Download the release assets, replacing the repository and version as needed:

```sh
gh release download v2.0.0 --repo mohuganda/CLIMS \
  --pattern images.env --pattern images.json \
  --dir artifacts/release/v2.0.0

docker compose --env-file infra/env/prod.env \
  --env-file artifacts/release/v2.0.0/images.env \
  -f infra/compose/docker-compose.yml \
  -f infra/compose/docker-compose.prod.yml build postgres nginx

docker compose --env-file infra/env/prod.env \
  --env-file artifacts/release/v2.0.0/images.env \
  -f infra/compose/docker-compose.yml \
  -f infra/compose/docker-compose.prod.yml pull backend frontend

docker compose --env-file infra/env/prod.env \
  --env-file artifacts/release/v2.0.0/images.env \
  -f infra/compose/docker-compose.yml \
  -f infra/compose/docker-compose.prod.yml up -d --no-build --wait
```

Use `docker login ghcr.io` with credentials authorized for the packages when pulling private images. Keep both env files and Compose files on subsequent commands. The second env file supplies image references; the first supplies database/JWT settings. PostgreSQL data remains in the production volume.

To use the released mobile exporter, use the same Compose options with `--profile artifacts pull mobile`, then `run --rm --no-deps mobile`. Its bundles are written to `artifacts/mobile/mobile-bundles.tar.gz`. The published mobile API URL is already baked in; local env overrides do not change it.

Without image overrides, production Compose retains the local `clims-<service>:prod` names and existing build commands. CI checks both the defaults and digest override behavior. Release publication does not deploy or restart any environment.

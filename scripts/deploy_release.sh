#!/usr/bin/env bash
# Runs on the production Linux server after a release bundle has been uploaded.
set -euo pipefail

root=${1:?Usage: deploy_release.sh /absolute/deployment/path}
release=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
[[ $root = /* && $root != / ]]
test -s "$root/shared/prod.env"
test -s "$release/images.env"

# Also serialize manual recovery commands with workflow deployments.
exec 9> "$root/deploy.lock"
flock -w 900 9
compose=(docker compose --project-name clims-prod
  --env-file "$root/shared/prod.env" --env-file "$release/images.env"
  -f "$release/infra/compose/docker-compose.yml"
  -f "$release/infra/compose/docker-compose.prod.yml")

# Fail before changing running services if configuration or image access is invalid.
"${compose[@]}" config --quiet
"${compose[@]}" pull backend frontend
"${compose[@]}" build postgres nginx
"${compose[@]}" up -d --no-build --wait --wait-timeout 300 postgres backend frontend nginx

# Keep the last healthy release discoverable; failed deployments do not advance it.
if [[ -L $root/current && $(readlink "$root/current") != "$release" ]]; then
  ln -sfn "$(readlink "$root/current")" "$root/previous"
fi
ln -sfn "$release" "$root/current"
echo "Deployment healthy: $release"

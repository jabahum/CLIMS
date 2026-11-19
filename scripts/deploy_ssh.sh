#!/usr/bin/env bash
set -euo pipefail

for name in DEPLOY_HOST DEPLOY_USER DEPLOY_SSH_KEY DEPLOY_KNOWN_HOSTS DEPLOY_PATH DEPLOY_PORT; do
  if [[ -z ${!name:-} ]]; then
    echo "Set production environment configuration: $name" >&2
    exit 1
  fi
done
[[ $DEPLOY_HOST =~ ^[a-zA-Z0-9][a-zA-Z0-9.-]*$ ]]
[[ $DEPLOY_USER =~ ^[a-zA-Z_][a-zA-Z0-9_-]*$ ]]
[[ $DEPLOY_PORT =~ ^[0-9]+$ ]] && (( DEPLOY_PORT > 0 && DEPLOY_PORT < 65536 ))
[[ $DEPLOY_PATH =~ ^/[a-zA-Z0-9_/-]+$ && $DEPLOY_PATH != / ]]

temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT
chmod 700 "$temporary"
printf '%s\n' "$DEPLOY_SSH_KEY" > "$temporary/key"
printf '%s\n' "$DEPLOY_KNOWN_HOSTS" > "$temporary/known_hosts"
chmod 600 "$temporary/key" "$temporary/known_hosts"
ssh_options=(-i "$temporary/key" -o IdentitiesOnly=yes -o BatchMode=yes
  -o StrictHostKeyChecking=yes -o "UserKnownHostsFile=$temporary/known_hosts"
  -o ConnectTimeout=15 -o ServerAliveInterval=15 -o ServerAliveCountMax=3)
target="$DEPLOY_USER@$DEPLOY_HOST"
destination="$DEPLOY_PATH/releases/$GITHUB_RUN_ID-$GITHUB_RUN_ATTEMPT"

# Include only tracked runtime configuration from this release, never local env files.
git archive HEAD infra/compose infra/docker/nginx infra/docker/postgres infra/nginx scripts/deploy_release.sh |
  tar -x -C "$temporary"
cp release/images.env "$temporary/images.env"
tar -czf "$temporary/release.tar.gz" -C "$temporary" infra scripts images.env
printf -v command 'mkdir -p %q && tar -xzf - -C %q' "$destination" "$destination"
ssh "${ssh_options[@]}" -p "$DEPLOY_PORT" "$target" "$command" < "$temporary/release.tar.gz"
printf -v command 'bash %q %q' "$destination/scripts/deploy_release.sh" "$DEPLOY_PATH"
ssh "${ssh_options[@]}" -p "$DEPLOY_PORT" "$target" "$command"

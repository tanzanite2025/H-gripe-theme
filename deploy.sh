#!/usr/bin/env bash
set -euo pipefail

TARGET_REMOTE="origin"
TARGET_BRANCH="master"
COMPOSE_FILE="compose.prod.yml"
ENV_FILE="deployment/production.env"
EDGE_NETWORK="tanzanite-edge"
PULL_ATTEMPTS="${PULL_ATTEMPTS:-60}"
PULL_DELAY_SECONDS="${PULL_DELAY_SECONDS:-15}"

cd "$(dirname "$0")"

for required_command in git docker; do
  if ! command -v "${required_command}" >/dev/null 2>&1; then
    echo "ERR: required command is not available: ${required_command}" >&2
    exit 1
  fi
done

if ! docker compose version >/dev/null 2>&1; then
  echo "ERR: Docker Compose v2 is required." >&2
  exit 1
fi

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "ERR: missing ${ENV_FILE}. Copy deployment/production.env.example first." >&2
  exit 1
fi

if grep -Eq '^[A-Za-z_][A-Za-z0-9_]*=CHANGE_ME' "${ENV_FILE}"; then
  echo "ERR: ${ENV_FILE} still contains CHANGE_ME placeholders." >&2
  exit 1
fi

if [[ ! "${PULL_ATTEMPTS}" =~ ^[1-9][0-9]*$ ]] || [[ ! "${PULL_DELAY_SECONDS}" =~ ^[1-9][0-9]*$ ]]; then
  echo "ERR: PULL_ATTEMPTS and PULL_DELAY_SECONDS must be positive integers." >&2
  exit 1
fi

if ! docker network inspect "${EDGE_NETWORK}" >/dev/null 2>&1; then
  echo "ERR: shared edge network ${EDGE_NETWORK} does not exist." >&2
  echo "Deploy the shared tanzanite-edge gateway before this project." >&2
  exit 1
fi

git fetch --all --prune
current_branch="$(git rev-parse --abbrev-ref HEAD)"
if [[ "${current_branch}" != "${TARGET_BRANCH}" ]]; then
  if git show-ref --verify --quiet "refs/heads/${TARGET_BRANCH}"; then
    git checkout "${TARGET_BRANCH}"
  else
    git checkout -b "${TARGET_BRANCH}" "${TARGET_REMOTE}/${TARGET_BRANCH}"
  fi
fi

deploy_ref="${DEPLOY_REF:-${TARGET_REMOTE}/${TARGET_BRANCH}}"
release_sha="$(git rev-parse "${deploy_ref}^{commit}")"
if [[ ! "${release_sha}" =~ ^[0-9a-f]{40}$ ]]; then
  echo "ERR: ${deploy_ref} did not resolve to a full Git commit SHA." >&2
  exit 1
fi

git reset --hard "${release_sha}"
git clean -fd -e deployment/production.env

export IMAGE_TAG="sha-${release_sha}"
echo "Deploying ${IMAGE_TAG} from ${deploy_ref}."

compose=(docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}")
"${compose[@]}" config --quiet

pull_succeeded=false
for ((attempt = 1; attempt <= PULL_ATTEMPTS; attempt++)); do
  echo "Pulling release images (attempt ${attempt}/${PULL_ATTEMPTS})..."
  if "${compose[@]}" pull; then
    pull_succeeded=true
    break
  fi
  if ((attempt < PULL_ATTEMPTS)); then
    sleep "${PULL_DELAY_SECONDS}"
  fi
done

if [[ "${pull_succeeded}" != "true" ]]; then
  echo "ERR: release images for ${IMAGE_TAG} were not available." >&2
  exit 1
fi

"${compose[@]}" up -d --remove-orphans
"${compose[@]}" ps

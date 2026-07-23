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

env_value() {
  local key="$1"
  local line
  line="$(grep -E "^[[:space:]]*${key}=" "${ENV_FILE}" | tail -n 1 || true)"
  line="${line#*=}"
  line="${line%$'\r'}"
  printf '%s' "${line}"
}

require_env_key() {
  local key="$1"
  if ! grep -Eq "^[[:space:]]*${key}=" "${ENV_FILE}"; then
    echo "ERR: ${ENV_FILE} is missing ${key}." >&2
    exit 1
  fi
}

require_positive_int_env() {
  local key="$1"
  local value
  require_env_key "${key}"
  value="$(env_value "${key}")"
  if [[ ! "${value}" =~ ^[1-9][0-9]*$ ]]; then
    echo "ERR: ${key} must be a positive integer." >&2
    exit 1
  fi
}

require_non_negative_int_env() {
  local key="$1"
  local value
  require_env_key "${key}"
  value="$(env_value "${key}")"
  if [[ ! "${value}" =~ ^[0-9]+$ ]]; then
    echo "ERR: ${key} must be a non-negative integer." >&2
    exit 1
  fi
}

for required_env_key in \
  REDIS_PASSWORD \
  JWT_SECRET \
  NUXT_HTML_CACHE_DRIVER \
  NUXT_HTML_CACHE_PREFIX \
  NUXT_HTML_CACHE_REDIS_DB \
  NUXT_HTML_CACHE_REDIS_TTL_SECONDS \
  NUXT_HTML_CACHE_REDIS_SCAN_COUNT \
  NUXT_HTML_CACHE_PURGE_TOKEN \
  STOREFRONT_HTML_CACHE_PURGE_DEBOUNCE_MS; do
  require_env_key "${required_env_key}"
done

html_cache_driver="$(env_value NUXT_HTML_CACHE_DRIVER)"
if [[ "${html_cache_driver}" != "redis" ]]; then
  echo "ERR: NUXT_HTML_CACHE_DRIVER must be redis in production." >&2
  exit 1
fi

html_cache_prefix="$(env_value NUXT_HTML_CACHE_PREFIX)"
if [[ -z "${html_cache_prefix}" ]]; then
  echo "ERR: NUXT_HTML_CACHE_PREFIX must not be empty." >&2
  exit 1
fi

require_non_negative_int_env NUXT_HTML_CACHE_REDIS_DB
require_positive_int_env NUXT_HTML_CACHE_REDIS_TTL_SECONDS
require_positive_int_env NUXT_HTML_CACHE_REDIS_SCAN_COUNT
require_positive_int_env STOREFRONT_HTML_CACHE_PURGE_DEBOUNCE_MS

html_cache_purge_token="$(env_value NUXT_HTML_CACHE_PURGE_TOKEN)"
if (( ${#html_cache_purge_token} < 32 )); then
  echo "ERR: NUXT_HTML_CACHE_PURGE_TOKEN must be at least 32 characters." >&2
  exit 1
fi

if [[ "${html_cache_purge_token}" == "$(env_value REDIS_PASSWORD)" ]] || [[ "${html_cache_purge_token}" == "$(env_value JWT_SECRET)" ]]; then
  echo "ERR: NUXT_HTML_CACHE_PURGE_TOKEN must be unique and must not reuse REDIS_PASSWORD or JWT_SECRET." >&2
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

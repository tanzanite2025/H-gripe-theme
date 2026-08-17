#!/usr/bin/env bash
set -euo pipefail

TARGET_REMOTE="origin"
TARGET_BRANCH="master"
COMPOSE_FILE="compose.prod.yml"
ENV_FILE="deployment/production.env"
EDGE_NETWORK="shared-edge"
PULL_ATTEMPTS="${PULL_ATTEMPTS:-60}"
PULL_DELAY_SECONDS="${PULL_DELAY_SECONDS:-15}"

cd "$(dirname "$0")"

for required_command in git docker; do
  if ! command -v "${required_command}" >/dev/null 2>&1; then
    echo "ERR: required command is not available: ${required_command}" >&2
    exit 1
  fi
done

PYTHON_CMD=()
for python_candidate in python3 python; do
  if command -v "${python_candidate}" >/dev/null 2>&1 &&
    "${python_candidate}" --version >/dev/null 2>&1; then
    PYTHON_CMD=("${python_candidate}")
    break
  fi
done
if [[ "${#PYTHON_CMD[@]}" -eq 0 ]] &&
  command -v py >/dev/null 2>&1 &&
  py -3 --version >/dev/null 2>&1; then
  PYTHON_CMD=(py -3)
fi
if [[ "${#PYTHON_CMD[@]}" -eq 0 ]]; then
  echo "ERR: a working Python 3 interpreter is required (python3, python, or py -3)." >&2
  exit 1
fi

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
  PAYMENT_CONFIG_MASTER_KEY \
  OPS_CONNECTOR_MASTER_KEY \
  OPS_EDGE_CONFIG_ENVIRONMENT \
  OPS_EDGE_CONFIG_DIR \
  EDGE_GATEWAY_ROUTE_FILE \
  EDGE_GATEWAY_CONTAINER \
  EDGE_GATEWAY_CADDYFILE \
  SERVER_BASE_URL \
  STOREFRONT_BASE_URL \
  NUXT_PUBLIC_SITE_URL \
  CORS_ORIGINS \
  TURNSTILE_REQUIRED \
  TURNSTILE_SECRET_KEY \
  TURNSTILE_SITE_KEY \
  VERIFICATION_IP_WINDOW_SECONDS \
  VERIFICATION_DESTINATION_WINDOW_SECONDS \
  VERIFICATION_DAILY_LIMIT \
  VERIFICATION_GLOBAL_WINDOW_SECONDS \
  VERIFICATION_GLOBAL_LIMIT \
  VERIFICATION_CIRCUIT_SECONDS \
  PAYMENT_RISK_FAILURE_WINDOW_SECONDS \
  PAYMENT_RISK_FAILURE_THRESHOLD \
  PAYMENT_RISK_DELAY_SECONDS \
  PAYMENT_RISK_HIGH_RISK_SCORE \
  REQUEST_SIGNING_ENABLED \
  REQUEST_SIGNING_KEY \
  REQUEST_SIGNING_MAX_SKEW_SECONDS \
  REQUEST_SIGNING_REQUIRED_PATHS \
  SMTP_HOST \
  SMTP_PORT \
  SMTP_USERNAME \
  SMTP_PASSWORD \
  SMTP_FROM \
  SMTP_FROM_NAME \
  NUXT_HTML_CACHE_DRIVER \
  NUXT_HTML_CACHE_PREFIX \
  NUXT_HTML_CACHE_REDIS_DB \
  NUXT_HTML_CACHE_REDIS_TTL_SECONDS \
  NUXT_HTML_CACHE_REDIS_SCAN_COUNT \
  NUXT_HTML_CACHE_PURGE_TOKEN \
  STOREFRONT_HTML_CACHE_PURGE_DEBOUNCE_MS; do
  require_env_key "${required_env_key}"
done

for required_secret in JWT_SECRET PAYMENT_CONFIG_MASTER_KEY OPS_CONNECTOR_MASTER_KEY TURNSTILE_SECRET_KEY REQUEST_SIGNING_KEY NUXT_HTML_CACHE_PURGE_TOKEN; do
  secret_value="$(env_value "${required_secret}")"
  if (( ${#secret_value} < 32 )); then
    echo "ERR: ${required_secret} must be at least 32 characters." >&2
    exit 1
  fi
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

site_quality_enabled="$(env_value WORKER_SITE_QUALITY_ENABLED)"
case "${site_quality_enabled}" in
  true|false)
    ;;
  *)
    echo "ERR: WORKER_SITE_QUALITY_ENABLED must be true or false." >&2
    exit 1
    ;;
esac

if [[ "${site_quality_enabled}" == "true" ]]; then
  require_env_key SITE_QUALITY_RUNNER_TOKEN
  site_quality_runner_token="$(env_value SITE_QUALITY_RUNNER_TOKEN)"
  if [[ -z "${site_quality_runner_token}" ]]; then
    echo "ERR: SITE_QUALITY_RUNNER_TOKEN is required when site quality monitoring is enabled." >&2
    exit 1
  fi
  if [[ "${#site_quality_runner_token}" -lt 32 ]]; then
    echo "ERR: SITE_QUALITY_RUNNER_TOKEN must be at least 32 characters." >&2
    exit 1
  fi
  require_positive_int_env WORKER_SITE_QUALITY_DISPATCH_INTERVAL_SECONDS
  require_positive_int_env WORKER_SITE_QUALITY_BATCH_LIMIT
  require_positive_int_env WORKER_SITE_QUALITY_LEASE_TIMEOUT_SECONDS
  require_positive_int_env WORKER_SITE_QUALITY_SAMPLE_COUNT
  require_positive_int_env WORKER_SITE_QUALITY_CONFIRMATIONS
  require_positive_int_env WORKER_SITE_QUALITY_CLEAN_EVALUATIONS
  require_positive_int_env WORKER_SITE_QUALITY_PROVIDER_CONCURRENCY
  require_non_negative_int_env WORKER_SITE_QUALITY_PROVIDER_SPACING_SECONDS
  if (( $(env_value WORKER_SITE_QUALITY_CONFIRMATIONS) > $(env_value WORKER_SITE_QUALITY_SAMPLE_COUNT) )); then
    echo "ERR: WORKER_SITE_QUALITY_CONFIRMATIONS must not exceed WORKER_SITE_QUALITY_SAMPLE_COUNT." >&2
    exit 1
  fi
fi

html_cache_purge_token="$(env_value NUXT_HTML_CACHE_PURGE_TOKEN)"
if (( ${#html_cache_purge_token} < 32 )); then
  echo "ERR: NUXT_HTML_CACHE_PURGE_TOKEN must be at least 32 characters." >&2
  exit 1
fi

if [[ "${html_cache_purge_token}" == "$(env_value REDIS_PASSWORD)" ]] || [[ "${html_cache_purge_token}" == "$(env_value JWT_SECRET)" ]]; then
  echo "ERR: NUXT_HTML_CACHE_PURGE_TOKEN must be unique and must not reuse REDIS_PASSWORD or JWT_SECRET." >&2
  exit 1
fi

edge_environment="$(env_value OPS_EDGE_CONFIG_ENVIRONMENT)"
edge_config_dir="$(env_value OPS_EDGE_CONFIG_DIR)"
edge_gateway_route_file="$(env_value EDGE_GATEWAY_ROUTE_FILE)"
edge_gateway_container="$(env_value EDGE_GATEWAY_CONTAINER)"
edge_gateway_caddyfile="$(env_value EDGE_GATEWAY_CADDYFILE)"

if [[ "${edge_environment}" != "production" ]]; then
  echo "ERR: OPS_EDGE_CONFIG_ENVIRONMENT must be production for deploy.sh." >&2
  exit 1
fi
if [[ "${edge_config_dir}" != /* ]]; then
  echo "ERR: OPS_EDGE_CONFIG_DIR must be an absolute host path." >&2
  exit 1
fi
if [[ "${edge_gateway_route_file}" != /* ]]; then
  echo "ERR: EDGE_GATEWAY_ROUTE_FILE must be an absolute host path." >&2
  exit 1
fi
if [[ -z "${edge_gateway_container}" || -z "${edge_gateway_caddyfile}" ]]; then
  echo "ERR: EDGE_GATEWAY_CONTAINER and EDGE_GATEWAY_CADDYFILE must not be empty." >&2
  exit 1
fi
if [[ "${edge_gateway_caddyfile}" != /* ]]; then
  echo "ERR: EDGE_GATEWAY_CADDYFILE must be an absolute container path." >&2
  exit 1
fi
if [[ "$(basename "${edge_gateway_route_file}")" != "commerce-platform.caddy" ]]; then
  echo "ERR: EDGE_GATEWAY_ROUTE_FILE must end with commerce-platform.caddy." >&2
  exit 1
fi

if [[ ! "${PULL_ATTEMPTS}" =~ ^[1-9][0-9]*$ ]] || [[ ! "${PULL_DELAY_SECONDS}" =~ ^[1-9][0-9]*$ ]]; then
  echo "ERR: PULL_ATTEMPTS and PULL_DELAY_SECONDS must be positive integers." >&2
  exit 1
fi

if ! docker network inspect "${EDGE_NETWORK}" >/dev/null 2>&1; then
  echo "ERR: shared edge network ${EDGE_NETWORK} does not exist." >&2
  echo "Deploy the shared-edge gateway before this project." >&2
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
git clean -fd -e deployment/production.env -e 'deployment/production.env.bak-*'

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

echo "Starting database and Redis dependencies..."
"${compose[@]}" up -d db redis

wait_for_service_health() {
  local service="$1"
  local attempts="${2:-60}"
  local container_id
  local health

  for ((attempt = 1; attempt <= attempts; attempt++)); do
    container_id="$("${compose[@]}" ps -q "${service}")"
    if [[ -n "${container_id}" ]]; then
      health="$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}no-healthcheck{{end}}' "${container_id}" 2>/dev/null || true)"
      if [[ "${health}" == "healthy" ]]; then
        return 0
      fi
    fi
    if (( attempt < attempts )); then
      sleep 2
    fi
  done

  echo "ERR: service ${service} did not become healthy." >&2
  "${compose[@]}" ps "${service}" >&2 || true
  return 1
}

wait_for_service_health db
wait_for_service_health redis

echo "Running database migrations..."
"${compose[@]}" up --no-deps --force-recreate --abort-on-container-exit --exit-code-from migrate migrate

echo "Backfilling persistent image derivatives for legacy media..."
"${compose[@]}" up --no-deps --force-recreate --abort-on-container-exit --exit-code-from media-derivatives-backfill media-derivatives-backfill

echo "Rendering edge configuration from the migrated database..."
"${compose[@]}" up --no-deps --force-recreate --abort-on-container-exit --exit-code-from edge-config edge-config

validate_generated_edge_config() {
  local manifest_path="${edge_config_dir}/manifest.json"
  local caddy_path="${edge_config_dir}/caddy.caddy"
  local storefront_names_path="${edge_config_dir}/storefront-server-names.conf"
  local admin_names_path="${edge_config_dir}/admin-server-names.conf"

  for artifact in "${manifest_path}" "${caddy_path}" "${storefront_names_path}" "${admin_names_path}"; do
    if [[ ! -s "${artifact}" ]]; then
      echo "ERR: edge-config did not produce a non-empty required artifact: ${artifact}" >&2
      return 1
    fi
  done

  "${PYTHON_CMD[@]}" - \
    "${manifest_path}" \
    "${caddy_path}" \
    "$(env_value SERVER_BASE_URL)" \
    "$(env_value STOREFRONT_BASE_URL)" \
    "$(env_value NUXT_PUBLIC_SITE_URL)" \
    "$(env_value CORS_ORIGINS)" \
    "$(env_value GOOGLE_MERCHANT_REDIRECT_URL)" \
    "$(env_value GOOGLE_MERCHANT_POST_CONNECT_URL)" <<'PY'
import json
import sys
from urllib.parse import urlparse

(
    manifest_path,
    caddy_path,
    server_base,
    storefront_base,
    site_url,
    cors_origins,
    google_merchant_redirect,
    google_merchant_post_connect,
) = sys.argv[1:]

with open(manifest_path, "r", encoding="utf-8") as fh:
    manifest = json.load(fh)
with open(caddy_path, "r", encoding="utf-8") as fh:
    caddy = fh.read()

canonical = str(manifest.get("canonical", "")).strip()
domains = manifest.get("domains") or []
nginx = manifest.get("nginx") or {}
storefront_names = set(nginx.get("storefront_server_names") or [])
admin_names = set(nginx.get("admin_server_names") or [])

errors = []
if not canonical:
    errors.append("manifest canonical domain is empty")
if not domains:
    errors.append("manifest has no generated domain routes")
if not caddy.strip():
    errors.append("generated Caddy route is empty")
if canonical not in storefront_names:
    errors.append("canonical domain is missing from generated storefront Nginx names")
if not admin_names:
    errors.append("generated admin Nginx names are empty")

expected_canonical_url = f"https://{canonical}"
for label, value in (
    ("SERVER_BASE_URL", server_base),
    ("STOREFRONT_BASE_URL", storefront_base),
    ("NUXT_PUBLIC_SITE_URL", site_url),
):
    if value.rstrip("/") != expected_canonical_url:
        errors.append(f"{label} must equal generated canonical URL {expected_canonical_url}")

expected_google_merchant_redirect = expected_canonical_url + "/api/admin/google-merchant/oauth/callback"
if google_merchant_redirect and google_merchant_redirect.rstrip("/") != expected_google_merchant_redirect:
    errors.append(
        "GOOGLE_MERCHANT_REDIRECT_URL must match the generated canonical URL callback path"
    )
expected_google_merchant_post_connect = expected_canonical_url + "/google-merchant"
if google_merchant_post_connect and google_merchant_post_connect.rstrip("/") != expected_google_merchant_post_connect:
    errors.append(
        "GOOGLE_MERCHANT_POST_CONNECT_URL must match the generated canonical URL"
    )

configured_cors = {item.strip().rstrip("/") for item in cors_origins.split(",") if item.strip()}
for domain in sorted(storefront_names | admin_names):
    if f"https://{domain}" not in configured_cors:
        errors.append(f"CORS_ORIGINS is missing generated domain https://{domain}")

for route in domains:
    domain = str(route.get("domain", "")).strip()
    if not domain or domain not in caddy:
        errors.append(f"generated Caddy route is missing domain {domain or '<empty>'}")

if errors:
    for error in errors:
        print(f"ERR: {error}", file=sys.stderr)
    raise SystemExit(1)

print(f"Generated edge config is valid for canonical domain {canonical}.")
PY
}

validate_generated_edge_config

echo "Starting application services after successful edge rendering..."
"${compose[@]}" up -d --remove-orphans api storefront admin web
wait_for_service_health api
wait_for_service_health storefront
wait_for_service_health admin
wait_for_service_health web

publish_edge_route() {
  local candidate="${edge_config_dir}/caddy.caddy"
  local route_file="${edge_gateway_route_file}"
  local route_dir
  local staged_route
  local backup_route
  local had_previous_route=false

  if ! docker inspect --format '{{.State.Running}}' "${edge_gateway_container}" 2>/dev/null | grep -qx true; then
    echo "ERR: shared edge gateway container is not running: ${edge_gateway_container}" >&2
    return 1
  fi

  route_dir="$(dirname "${route_file}")"
  mkdir -p "${route_dir}"
  staged_route="${route_file}.new.$$"
  backup_route="${route_file}.previous.$$"

  install -m 0644 "${candidate}" "${staged_route}"

  echo "Validating generated Caddy route before publishing..."
  if ! docker exec -i "${edge_gateway_container}" caddy adapt --config - --adapter caddyfile < "${staged_route}" >/dev/null; then
    rm -f -- "${staged_route}"
    echo "ERR: generated Caddy route failed standalone validation; existing gateway route was not changed." >&2
    return 1
  fi

  if [[ -e "${route_file}" ]]; then
    cp -- "${route_file}" "${backup_route}"
    had_previous_route=true
  fi

  mv -f -- "${staged_route}" "${route_file}"

  restore_previous_edge_route() {
    if [[ "${had_previous_route}" == "true" ]]; then
      mv -f -- "${backup_route}" "${route_file}"
    else
      rm -f -- "${route_file}"
    fi
  }

  echo "Validating shared edge Caddyfile with the new route..."
  if ! docker exec "${edge_gateway_container}" caddy validate --config "${edge_gateway_caddyfile}" --adapter caddyfile; then
    restore_previous_edge_route
    echo "ERR: shared edge Caddyfile rejected the generated route; existing gateway route was restored." >&2
    return 1
  fi

  echo "Reloading shared edge gateway..."
  if ! docker exec "${edge_gateway_container}" caddy reload --config "${edge_gateway_caddyfile}" --adapter caddyfile; then
    restore_previous_edge_route
    if ! docker exec "${edge_gateway_container}" caddy reload --config "${edge_gateway_caddyfile}" --adapter caddyfile >/dev/null 2>&1; then
      echo "ERR: gateway reload failed and restoring the previous route also failed." >&2
    else
      echo "ERR: gateway reload failed; existing gateway route was restored." >&2
    fi
    return 1
  fi

  rm -f -- "${backup_route}"
  echo "Shared edge route published and reloaded."
}

publish_edge_route

purge_storefront_html_cache() {
  local attempt
  for attempt in {1..15}; do
    if "${compose[@]}" exec -T storefront sh -lc \
      'wget -qO- --header="x-html-cache-purge-token: $NUXT_HTML_CACHE_PURGE_TOKEN" --header="content-type: application/json" --post-data="{}" http://127.0.0.1:3000/_internal/html-cache/purge' \
      >/dev/null 2>&1; then
      echo "Storefront HTML cache purged."
      return 0
    fi

    if (( attempt < 15 )); then
      sleep 2
    fi
  done

  echo "ERR: failed to purge storefront HTML cache after deployment." >&2
  return 1
}

purge_storefront_html_cache
"${compose[@]}" ps

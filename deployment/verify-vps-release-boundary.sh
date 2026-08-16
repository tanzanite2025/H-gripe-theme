#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-${ROOT_DIR}/compose.prod.yml}"
ENV_FILE="${ENV_FILE:-${ROOT_DIR}/deployment/production.env}"
REPORT_DIR="${REPORT_DIR:-${ROOT_DIR}/release-evidence/$(date -u +"%Y%m%dT%H%M%SZ")}"
CHECK_CONNECTIVITY="${CHECK_CONNECTIVITY:-false}"

log() {
  printf '[INFO] %s\n' "$*"
}

err() {
  printf '[ERROR] %s\n' "$*" >&2
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    err "required command is not available: $1"
    exit 1
  fi
}

if [[ ! -f "${ENV_FILE}" ]]; then
  if [[ "${ALLOW_EXAMPLE_ENV:-false}" == "true" && -f "${ROOT_DIR}/deployment/production.env.example" ]]; then
    ENV_FILE="${ROOT_DIR}/deployment/production.env.example"
    log "using deployment/production.env.example because ALLOW_EXAMPLE_ENV=true"
  else
    err "missing ${ENV_FILE}; set ENV_FILE or create deployment/production.env"
    exit 1
  fi
fi

require_command docker

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
  err "a working Python 3 interpreter is required (python3, python, or py -3)"
  exit 1
fi

if ! docker compose version >/dev/null 2>&1; then
  err "Docker Compose v2 is required"
  exit 1
fi

if ! grep -Fq '"${compose[@]}" up --no-deps --force-recreate --abort-on-container-exit --exit-code-from migrate migrate' "${ROOT_DIR}/deploy.sh"; then
  err "deploy.sh must force-recreate and run the migrate job before application services"
  exit 1
fi

legacy_public_patterns='tanzanite\.site|www\.tanzanite\.site|admin\.tanzanite\.site|tanzanite-edge'
legacy_public_files=(
  "${ROOT_DIR}/compose.prod.yml"
  "${ROOT_DIR}/deploy.sh"
  "${ROOT_DIR}/deployment/production.env.example"
  "${ROOT_DIR}/deployment/nginx/theme-web.conf"
  "${ROOT_DIR}/deployment/edge/commerce-platform.caddy"
  "${ROOT_DIR}/go-backend/config/config.production.yaml"
  "${ROOT_DIR}/go-backend/config/config.example.yaml"
  "${ROOT_DIR}/go-backend/.env.example"
  "${ROOT_DIR}/go-backend/cmd/server/swagger.go"
  "${ROOT_DIR}/go-backend/docs/swagger.yaml"
  "${ROOT_DIR}/go-backend/frontend-examples/nuxt3/nuxt.config.example.ts"
  "${ROOT_DIR}/go-backend/web/admin/src/views/GoogleMerchant.vue"
)

for file in "${legacy_public_files[@]}"; do
  if grep -nE "${legacy_public_patterns}" "${file}" >/dev/null 2>&1; then
    err "legacy public brand reference still present in ${file}"
    grep -nE "${legacy_public_patterns}" "${file}" >&2 || true
    exit 1
  fi
done

mkdir -p "${REPORT_DIR}"
compose=(docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}")

log "writing sanitized Compose JSON evidence to ${REPORT_DIR}"
raw_compose_json="$(mktemp)"
trap 'rm -f "${raw_compose_json}"' EXIT
"${compose[@]}" config --format json > "${raw_compose_json}"
"${PYTHON_CMD[@]}" - "${raw_compose_json}" "${REPORT_DIR}/compose-config.json" <<'PY'
import json
import re
import sys

source_path, evidence_path = sys.argv[1:3]
with open(source_path, "r", encoding="utf-8") as fh:
    data = json.load(fh)

sensitive_key = re.compile(
    r"(?:PASSWORD|SECRET|TOKEN|PRIVATE_KEY|API_KEY|PUBLISHABLE_KEY|CREDENTIAL|CERTIFICATE)",
    re.IGNORECASE,
)


def redact(value):
    if isinstance(value, dict):
        for key, child in value.items():
            if sensitive_key.search(str(key)):
                value[key] = "<redacted>"
            else:
                redact(child)
    elif isinstance(value, list):
        for child in value:
            redact(child)


redact(data)
with open(evidence_path, "w", encoding="utf-8") as fh:
    json.dump(data, fh, indent=2, sort_keys=True)
    fh.write("\n")
PY

resolve_network_name() {
  local logical_name="$1"
  "${PYTHON_CMD[@]}" - "${REPORT_DIR}/compose-config.json" "${logical_name}" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as fh:
    data = json.load(fh)

logical_name = sys.argv[2]
network = (data.get("networks") or {}).get(logical_name)
resolved_name = network.get("name") if isinstance(network, dict) else None
if not isinstance(resolved_name, str) or not resolved_name:
    raise SystemExit(f"missing resolved network name: {logical_name}")

print(resolved_name)
PY
}

db_network_name="$(resolve_network_name db)"
cache_network_name="$(resolve_network_name cache)"
app_network_name="$(resolve_network_name app)"
edge_network_name="$(resolve_network_name edge)"

"${PYTHON_CMD[@]}" - "${REPORT_DIR}/compose-config.json" <<'PY'
import json
import sys

config_path = sys.argv[1]
with open(config_path, "r", encoding="utf-8") as fh:
    data = json.load(fh)

errors = []
services = data.get("services", {})
networks = data.get("networks", {})

expected_service_networks = {
    "db": {"db"},
    "redis": {"cache"},
    "migrate": {"db"},
    "edge-config": {"db"},
    "api": {"db", "cache", "app"},
    "storefront": {"app", "cache"},
    "admin": {"app"},
    "web": {"app", "edge"},
}

for service_name, expected_networks in expected_service_networks.items():
    service = services.get(service_name)
    if not service:
        errors.append(f"missing service: {service_name}")
        continue

    if service.get("ports"):
        errors.append(f"{service_name} must not publish host ports")
    if service.get("container_name"):
        errors.append(f"{service_name} must not set container_name")
    if service.get("network_mode") == "host":
        errors.append(f"{service_name} must not use host network mode")

    actual_networks = service.get("networks") or {}
    if isinstance(actual_networks, dict):
        actual_networks = set(actual_networks.keys())
    else:
        actual_networks = set(actual_networks)
    if actual_networks != expected_networks:
        errors.append(
            f"{service_name} networks mismatch: expected {sorted(expected_networks)}, got {sorted(actual_networks)}"
        )

edge_config = services.get("edge-config") or {}
edge_config_environment = edge_config.get("environment") or {}
if isinstance(edge_config_environment, dict):
    for forbidden_key in ("REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD"):
        if forbidden_key in edge_config_environment:
            errors.append(f"edge-config must not depend on Redis environment: {forbidden_key}")
if set((edge_config.get("depends_on") or {}).keys()) != {"migrate"}:
    errors.append("edge-config must depend only on the completed migrate service")

migrate = services.get("migrate") or {}
if migrate.get("command") != ["migrate"]:
    errors.append("migrate must run only the migrate command")
if migrate.get("restart") != "no":
    errors.append("migrate must be a one-shot service with restart=no")
if migrate.get("pull_policy") != "always":
    errors.append("migrate must always pull the release API image")
migration_db_dependency = (migrate.get("depends_on") or {}).get("db") or {}
if migration_db_dependency.get("condition") != "service_healthy":
    errors.append("migrate must wait for db service_healthy")

for service_name in ("edge-config", "api", "storefront", "admin", "web"):
    dependency = ((services.get(service_name) or {}).get("depends_on") or {}).get("migrate") or {}
    if dependency.get("condition") != "service_completed_successfully":
        errors.append(f"{service_name} must wait for migrate service_completed_successfully")

web = services.get("web") or {}
edge_config_dependency = (web.get("depends_on") or {}).get("edge-config") or {}
if edge_config_dependency.get("condition") != "service_completed_successfully":
    errors.append("web must wait for edge-config service_completed_successfully")

generated_edge_mounts = [
    volume
    for volume in (web.get("volumes") or [])
    if isinstance(volume, dict) and volume.get("target") == "/etc/nginx/generated-edge"
]
if not generated_edge_mounts or any(volume.get("read_only") is not True for volume in generated_edge_mounts):
    errors.append("web must mount generated Nginx files read-only")

for network_name in ("db", "cache"):
    network = networks.get(network_name) or {}
    if not network.get("name"):
        errors.append(f"{network_name} network must have a resolved name")
    if network.get("internal") is not True:
        errors.append(f"{network_name} network must be internal")

edge = networks.get("edge") or {}
if not edge.get("name"):
    errors.append("edge network must have a resolved name")
if edge.get("external") is not True:
    errors.append("edge network must be external")
if edge.get("name") != "shared-edge":
    errors.append("edge network must be named shared-edge")

edge_members = [
    name
    for name, service in services.items()
    if "edge" in ((service.get("networks") or {}).keys() if isinstance(service.get("networks") or {}, dict) else service.get("networks") or [])
]
if edge_members != ["web"]:
    errors.append(f"only web may join edge network, got {edge_members}")

if errors:
    for error in errors:
        print(f"FAIL: {error}", file=sys.stderr)
    sys.exit(1)

print("OK: production Compose network boundary is statically valid")
print("OK: no business service publishes host ports in Compose config")
print("OK: db/cache are internal and only web joins the external edge network")
print(
    "OK: resolved Docker networks: "
    f"db={networks['db']['name']} "
    f"cache={networks['cache']['name']} "
    f"app={networks['app']['name']} "
    f"edge={networks['edge']['name']}"
)
PY

if [[ "${CHECK_CONNECTIVITY}" != "true" ]]; then
  log "static boundary checks complete; set CHECK_CONNECTIVITY=true on the VPS after deployment for runtime checks"
  exit 0
fi

log "collecting runtime Compose and Docker network evidence"
"${compose[@]}" ps --format json > "${REPORT_DIR}/compose-ps.json"

for network in "${db_network_name}" "${cache_network_name}" "${app_network_name}" "${edge_network_name}"; do
  docker network inspect "${network}" > "${REPORT_DIR}/network-${network}.json"
done

"${PYTHON_CMD[@]}" - "${REPORT_DIR}/compose-ps.json" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as fh:
    raw = fh.read().strip()

records = []
if raw:
    try:
        parsed = json.loads(raw)
        records = parsed if isinstance(parsed, list) else [parsed]
    except json.JSONDecodeError:
        records = [json.loads(line) for line in raw.splitlines() if line.strip()]

published = []
for record in records:
    for publisher in record.get("Publishers") or record.get("publishers") or []:
        published_port = publisher.get("PublishedPort") or publisher.get("published_port")
        if published_port:
            published.append((record.get("Service") or record.get("Name"), published_port))

if published:
    for service, port in published:
        print(f"FAIL: {service} publishes host port {port}", file=sys.stderr)
    sys.exit(1)

print("OK: running Compose services do not publish host ports")
PY

connectivity_log="${REPORT_DIR}/connectivity.txt"
: > "${connectivity_log}"

run_diagnostic() {
  local label="$1"
  local network="$2"
  local host="$3"
  local port="$4"
  local expect_success="$5"

  {
    printf '\n[%s]\n' "${label}"
    printf 'network=%s target=%s:%s expect_success=%s\n' "${network}" "${host}" "${port}" "${expect_success}"
  } >> "${connectivity_log}"

  set +e
  docker run --rm --network "${network}" alpine:3.20 sh -lc "nc -z -w 2 ${host} ${port}" >> "${connectivity_log}" 2>&1
  local status=$?
  set -e

  if [[ "${expect_success}" == "true" && "${status}" -ne 0 ]]; then
    err "${label} failed; see ${connectivity_log}"
    exit 1
  fi
  if [[ "${expect_success}" != "true" && "${status}" -eq 0 ]]; then
    err "${label} unexpectedly succeeded; see ${connectivity_log}"
    exit 1
  fi

  printf 'status=%s\n' "${status}" >> "${connectivity_log}"
  log "${label}: observed expected ${expect_success}"
}

run_diagnostic "api network reaches PostgreSQL" "${db_network_name}" "db" 5432 true
run_diagnostic "api network reaches Redis" "${cache_network_name}" "redis" 6379 true
run_diagnostic "app network cannot reach PostgreSQL from storefront side" "${app_network_name}" "db" 5432 false

log "runtime connectivity checks complete; evidence written to ${REPORT_DIR}"

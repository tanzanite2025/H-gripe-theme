#!/usr/bin/env bash
set -euo pipefail
umask 077

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-${ROOT_DIR}/compose.prod.yml}"
ENV_FILE="${ENV_FILE:-${ROOT_DIR}/deployment/production.env}"
REPORT_DIR="${REPORT_DIR:-${ROOT_DIR}/release-evidence/$(date -u +"%Y%m%dT%H%M%SZ")}"
CHECK_CONNECTIVITY="${CHECK_CONNECTIVITY:-false}"
CHECK_RUNTIME_METADATA="${CHECK_RUNTIME_METADATA:-false}"

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

for check_setting in CHECK_CONNECTIVITY CHECK_RUNTIME_METADATA; do
  case "${!check_setting}" in
    true|false) ;;
    *) err "${check_setting} must be true or false"; exit 1 ;;
  esac
done

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
if ! grep -Fq '"${compose[@]}" up --no-deps --force-recreate --abort-on-container-exit --exit-code-from media-derivatives-backfill media-derivatives-backfill' "${ROOT_DIR}/deploy.sh"; then
  err "deploy.sh must force-recreate and run the media derivative backfill before application services"
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
if [[ -e "${REPORT_DIR}/compose-config.json" || -L "${REPORT_DIR}/compose-config.json" ]]; then
  err "report directory already contains release evidence; select a fresh REPORT_DIR"
  exit 1
fi
source "${ROOT_DIR}/deployment/local-computer-backup-release-overlay-selection-and-existing-maintenance-gate-protection.sh"
local_backup_maintenance_enabled="$(sed -n 's/^[[:space:]]*LOCAL_BACKUP_MAINTENANCE_ENABLED=//p' "${ENV_FILE}" | tail -n 1)"
local_backup_maintenance_enabled="${local_backup_maintenance_enabled%$'\r'}"
local_backup_maintenance_enabled="${local_backup_maintenance_enabled:-false}"
redis_mode="$(sed -n 's/^[[:space:]]*REDIS_MODE=//p' "${ENV_FILE}" | tail -n 1)"
redis_mode="${redis_mode%$'\r'}"
compose=(docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}")
append_redis_sentinel_overlay_for_explicit_release_mode "${ROOT_DIR}" "${redis_mode}" compose
append_local_computer_backup_overlay_for_explicit_release_setting "${ROOT_DIR}" "${local_backup_maintenance_enabled}" compose

log "writing sanitized Compose JSON evidence to ${REPORT_DIR}"
raw_compose_json="$(mktemp)"
raw_compose_ps_json=""
raw_network_json=""
trap 'rm -f "${raw_compose_json}" "${raw_compose_ps_json}" "${raw_network_json}"' EXIT
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
api_ingress_network_name="$(resolve_network_name api_ingress)"
edge_network_name="$(resolve_network_name edge)"

"${PYTHON_CMD[@]}" - "${REPORT_DIR}/compose-config.json" <<'PY'
import json
import sys
import re

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
    "media-derivatives-backfill": {"db"},
    "edge-config": {"db"},
    "api": {"db", "cache", "app", "api_ingress"},
    "storefront": {"app", "cache"},
    "admin": {"app"},
    "web": {"app", "api_ingress", "edge"},
}
if "local-backup-maintenance" in services:
    expected_service_networks["local-backup-maintenance"] = {"db", "app"}
for service_name in ("redis-replica-1", "redis-replica-2", "redis-sentinel-1", "redis-sentinel-2", "redis-sentinel-3"):
    if service_name in services:
        expected_service_networks[service_name] = {"cache"}

for service_name, service in services.items():
    if service.get("ports") and service_name != "local-backup-maintenance":
        errors.append(f"{service_name} must not publish host ports")
    if service.get("container_name"):
        errors.append(f"{service_name} must not set container_name")
    if service.get("network_mode") == "host":
        errors.append(f"{service_name} must not use host network mode")

for service_name, expected_networks in expected_service_networks.items():
    service = services.get(service_name)
    if not service:
        errors.append(f"missing service: {service_name}")
        continue

    actual_networks = service.get("networks") or {}
    if isinstance(actual_networks, dict):
        actual_networks = set(actual_networks.keys())
    else:
        actual_networks = set(actual_networks)
    if actual_networks != expected_networks:
        errors.append(
            f"{service_name} networks mismatch: expected {sorted(expected_networks)}, got {sorted(actual_networks)}"
        )

maintenance = services.get("local-backup-maintenance")
if maintenance:
    ports = maintenance.get("ports") or []
    if len(ports) != 1 or ports[0].get("host_ip") != "127.0.0.1" or ports[0].get("target") != 19090 or ports[0].get("protocol", "tcp") != "tcp":
        errors.append("backup maintenance must publish only its loopback TCP endpoint")
    published_port = str(ports[0].get("published", "")) if len(ports) == 1 else ""
    if not published_port.isascii() or not published_port.isdigit() or not 1 <= int(published_port) <= 65535:
        errors.append("backup maintenance must select exactly one valid host port")
    if maintenance.get("depends_on"):
        errors.append("backup maintenance must remain independent of business service startup")
    if maintenance.get("read_only") is not True:
        errors.append("backup maintenance root filesystem must be read-only")
    storefront = services.get("storefront") or {}
    image_tags = [str(service.get("image", "")).rsplit(":", 1)[-1] for service in (maintenance, storefront, services.get("api") or {})]
    if len(set(image_tags)) != 1 or not re.fullmatch(r"sha-[a-f0-9]{40}", image_tags[0]):
        errors.append("backup and storefront static files require the same immutable release image tag")
    storefront_static = [mount for mount in storefront.get("volumes", []) if mount.get("target") == "/app/.output/public/company"]
    maintenance_static = [mount for mount in maintenance.get("volumes", []) if mount.get("target") == "/app/product-static-assets/company"]
    if len(storefront_static) != 1 or len(maintenance_static) != 1:
        errors.append("backup must share the running storefront company static volume")
    else:
        serving, recovery = storefront_static[0], maintenance_static[0]
        definition = (data.get("volumes") or {}).get(serving.get("source"), {})
        if serving.get("type") != "volume" or recovery.get("type") != "volume" or serving.get("source") != recovery.get("source") or not serving.get("read_only") or recovery.get("read_only"):
            errors.append("storefront static volume must be shared, read-only for serving and writable for paired recovery")
        if (serving.get("volume") or {}).get("nocopy") or not (recovery.get("volume") or {}).get("nocopy"):
            errors.append("only the storefront image may initialize its company static volume")
        if not str(definition.get("name", "")).endswith("-" + image_tags[0]):
            errors.append("storefront company static volume must be isolated by immutable release")
    gate_sources = []
    for name in ("api", "migrate", "media-derivatives-backfill", "local-backup-maintenance"):
        gate_mounts = [mount for mount in services[name].get("volumes", []) if mount.get("target") == "/app/local-backup-maintenance-gate"]
        if len(gate_mounts) != 1 or gate_mounts[0].get("type") != "bind" or gate_mounts[0].get("read_only"):
            errors.append(f"{name} must mount the writable shared backup maintenance gate")
        else:
            gate_sources.append(gate_mounts[0].get("source"))
    if len(gate_sources) != 4 or len(set(gate_sources)) != 1:
        errors.append("all backup-aware release services must use the same host maintenance gate")

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

media_backfill = services.get("media-derivatives-backfill") or {}
if media_backfill.get("command") != ["backfill-media-derivatives"]:
    errors.append("media-derivatives-backfill must run only the backfill-media-derivatives command")
if media_backfill.get("restart") != "no":
    errors.append("media-derivatives-backfill must be a one-shot service with restart=no")
if media_backfill.get("pull_policy") != "always":
    errors.append("media-derivatives-backfill must always pull the release API image")
backfill_dependency = (media_backfill.get("depends_on") or {}).get("migrate") or {}
if backfill_dependency.get("condition") != "service_completed_successfully":
    errors.append("media-derivatives-backfill must wait for migrate service_completed_successfully")

for service_name in ("edge-config", "api", "storefront", "admin", "web"):
    dependency = ((services.get(service_name) or {}).get("depends_on") or {}).get("migrate") or {}
    if dependency.get("condition") != "service_completed_successfully":
        errors.append(f"{service_name} must wait for migrate service_completed_successfully")

for service_name in ("api", "admin"):
    dependency = ((services.get(service_name) or {}).get("depends_on") or {}).get("media-derivatives-backfill") or {}
    if dependency.get("condition") != "service_completed_successfully":
        errors.append(f"{service_name} must wait for media-derivatives-backfill service_completed_successfully")

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

def service_network_endpoint(service, network_name):
    service_networks = service.get("networks") or {}
    if not isinstance(service_networks, dict):
        return {}
    endpoint = service_networks.get(network_name)
    return endpoint if isinstance(endpoint, dict) else {}


def environment_value(service, key):
    environment = service.get("environment") or {}
    if isinstance(environment, dict):
        return environment.get(key)
    prefix = f"{key}="
    for item in environment:
        if isinstance(item, str) and item.startswith(prefix):
            return item[len(prefix):]
    return None


api = services.get("api") or {}
if maintenance:
    for name in ("SERVER_BASE_URL", "STOREFRONT_BASE_URL"):
        if not environment_value(maintenance, name) or environment_value(maintenance, name) != environment_value(api, name):
            errors.append(f"backup maintenance must inherit the existing application {name}")
if environment_value(api, "MEDIA_REVIEWED_HISTORICAL_CUSTOMER_SERVICE_ATTACHMENT_ORIGINS") is None:
    errors.append("api must pass through MEDIA_REVIEWED_HISTORICAL_CUSTOMER_SERVICE_ATTACHMENT_ORIGINS, allowing an empty default")
web_ingress = service_network_endpoint(web, "api_ingress")
if web_ingress.get("ipv4_address") != "172.30.0.10":
    errors.append("web api_ingress address must be pinned to 172.30.0.10")

api_ingress_endpoint = service_network_endpoint(api, "api_ingress")
if "api-ingress" not in (api_ingress_endpoint.get("aliases") or []):
    errors.append("api must publish the api-ingress alias on api_ingress")

for service_name in ("api", "edge-config"):
    if environment_value(services.get(service_name) or {}, "TRUSTED_PROXIES") != "172.30.0.10/32":
        errors.append(f"{service_name} TRUSTED_PROXIES must be exactly 172.30.0.10/32")

for network_name in ("db", "cache", "api_ingress"):
    network = networks.get(network_name) or {}
    if not network.get("name"):
        errors.append(f"{network_name} network must have a resolved name")
    if network_name != "api_ingress" and network.get("internal") is not True:
        errors.append(f"{network_name} network must be internal")

api_ingress_network = networks.get("api_ingress") or {}
if api_ingress_network.get("internal") is not True:
    errors.append("api_ingress network must be internal")
api_ingress_subnets = {
    config.get("subnet")
    for config in ((api_ingress_network.get("ipam") or {}).get("config") or [])
    if isinstance(config, dict) and config.get("subnet")
}
if "172.30.0.0/24" not in api_ingress_subnets:
    errors.append("api_ingress network must use subnet 172.30.0.0/24")

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

api_ingress_members = [
    name
    for name, service in services.items()
    if "api_ingress" in ((service.get("networks") or {}).keys() if isinstance(service.get("networks") or {}, dict) else service.get("networks") or [])
]
if set(api_ingress_members) != {"api", "web"}:
    errors.append(f"only api and web may join api_ingress network, got {api_ingress_members}")

if errors:
    for error in errors:
        print(f"FAIL: {error}", file=sys.stderr)
    sys.exit(1)

print("OK: production Compose network boundary is statically valid")
print("OK: no business service publishes host ports in Compose config")
print("OK: db/cache are internal and only web joins the external edge network")
print("OK: api_ingress is an internal api-only ingress with a pinned web proxy address")
print(
    "OK: resolved Docker networks: "
    f"db={networks['db']['name']} "
    f"cache={networks['cache']['name']} "
    f"app={networks['app']['name']} "
    f"api_ingress={networks['api_ingress']['name']} "
    f"edge={networks['edge']['name']}"
)
PY

if [[ "${CHECK_CONNECTIVITY}" != "true" && "${CHECK_RUNTIME_METADATA}" != "true" ]]; then
  log "static boundary checks complete; use CHECK_RUNTIME_METADATA=true for read-only Docker metadata or CHECK_CONNECTIVITY=true for active network probes"
  exit 0
fi

log "collecting allowlisted runtime Compose and Docker network metadata"
raw_compose_ps_json="$(mktemp)"
raw_network_json="$(mktemp)"
"${compose[@]}" ps --all --format json > "${raw_compose_ps_json}"
docker network inspect "${db_network_name}" "${cache_network_name}" "${app_network_name}" "${api_ingress_network_name}" "${edge_network_name}" > "${raw_network_json}"

"${PYTHON_CMD[@]}" - "${REPORT_DIR}/compose-config.json" "${raw_compose_ps_json}" "${raw_network_json}" "${REPORT_DIR}" <<'PY'
import json
import pathlib
import re
import sys

config_path, ps_path, network_path, report_directory = sys.argv[1:5]
with open(config_path, "r", encoding="utf-8") as fh:
    configuration = json.load(fh)
with open(ps_path, "r", encoding="utf-8") as fh:
    raw = fh.read().strip()

records = []
if raw:
    try:
        parsed = json.loads(raw)
        records = parsed if isinstance(parsed, list) else [parsed]
    except json.JSONDecodeError:
        records = [json.loads(line) for line in raw.splitlines() if line.strip()]

errors = []
if not records or any(not isinstance(record, dict) for record in records):
    raise SystemExit("FAIL: runtime metadata must contain Compose container records")
services = configuration["services"]
maintenance = services.get("local-backup-maintenance")
allowed_maintenance_port = None
if maintenance:
    port = maintenance["ports"][0]
    allowed_maintenance_port = ("local-backup-maintenance", port["host_ip"], int(port["published"]), port["target"], port.get("protocol", "tcp"))


def read_runtime_metadata_field(record, field, alternate):
    return record.get(field, record.get(alternate))


def parse_runtime_port_number(value):
    if isinstance(value, bool) or not re.fullmatch(r"[0-9]+", str(value), flags=re.ASCII):
        return None
    port = int(value)
    return port if 1 <= port <= 65535 else None


sanitized_records = []
for record in records:
    service = read_runtime_metadata_field(record, "Service", "service")
    project = read_runtime_metadata_field(record, "Project", "project")
    if service not in services or project != configuration.get("name"):
        errors.append("runtime container does not belong to the selected project/service configuration")
    sanitized_record = {}
    for field in ("ID", "Name", "Image", "Project", "Service", "State", "Health", "ExitCode"):
        value = read_runtime_metadata_field(record, field, field.lower())
        if isinstance(value, (str, int)) and not isinstance(value, bool):
            sanitized_record[field] = value
    sanitized_publishers = []
    publishers = read_runtime_metadata_field(record, "Publishers", "publishers") or []
    if not isinstance(publishers, list):
        raise SystemExit("FAIL: invalid runtime port metadata")
    for publisher in publishers:
        if not isinstance(publisher, dict):
            raise SystemExit("FAIL: invalid runtime port binding")
        published_port = read_runtime_metadata_field(publisher, "PublishedPort", "published_port")
        if published_port is None or (not isinstance(published_port, bool) and published_port in (0, "0", "")):
            continue
        binding = (
            service,
            read_runtime_metadata_field(publisher, "URL", "url"),
            parse_runtime_port_number(published_port),
            parse_runtime_port_number(read_runtime_metadata_field(publisher, "TargetPort", "target_port")),
            read_runtime_metadata_field(publisher, "Protocol", "protocol"),
        )
        if allowed_maintenance_port is None or binding != allowed_maintenance_port:
            errors.append("runtime host port differs from the explicitly configured loopback maintenance endpoint")
        sanitized_publishers.append(dict(zip(("URL", "PublishedPort", "TargetPort", "Protocol"), binding[1:])))
    sanitized_record["Publishers"] = sanitized_publishers
    sanitized_records.append(sanitized_record)

with open(network_path, "r", encoding="utf-8") as fh:
    network_records = json.load(fh)
expected_networks = {configuration["networks"][name]["name"]: configuration["networks"][name] for name in ("db", "cache", "app", "api_ingress", "edge")}
sanitized_networks = {}
if not isinstance(network_records, list):
    raise SystemExit("FAIL: invalid Docker network metadata")
for network in network_records:
    if not isinstance(network, dict):
        raise SystemExit("FAIL: invalid Docker network record")
    name = network.get("Name")
    if not isinstance(name, str) or not re.fullmatch(r"[a-zA-Z0-9][a-zA-Z0-9_.-]*", name) or name not in expected_networks or name in sanitized_networks:
        errors.append("runtime network identity differs from selected Compose configuration")
        continue
    if network.get("Internal") is not bool(expected_networks[name].get("internal", False)):
        errors.append("runtime network internal boundary differs from selected Compose configuration")
    sanitized_networks[name] = {field: network[field] for field in ("Name", "Id", "Driver", "Scope", "Internal", "EnableIPv6") if field in network and isinstance(network[field], (str, bool))}
if set(sanitized_networks) != set(expected_networks):
    errors.append("runtime evidence is missing a selected Docker network")

if errors:
    for error in errors:
        print(f"FAIL: {error}", file=sys.stderr)
    sys.exit(1)


def write_allowlisted_runtime_metadata_evidence(filename, data):
    with (pathlib.Path(report_directory) / filename).open("w", encoding="utf-8") as fh:
        json.dump(data, fh, indent=2, sort_keys=True)
        fh.write("\n")


write_allowlisted_runtime_metadata_evidence("compose-ps.json", sanitized_records)
for name, network in sanitized_networks.items():
    write_allowlisted_runtime_metadata_evidence(f"network-{name}.json", [network])
write_allowlisted_runtime_metadata_evidence("runtime-metadata-boundary-summary.json", {
    "scope": "compose_container_and_network_metadata_only",
    "runtime_metadata_boundary_valid": True,
    "target_environment_acceptance_complete": False,
    "compose_project": configuration["name"],
    "container_count": len(sanitized_records),
    "network_count": len(sanitized_networks),
    "unverified_checks": ["container_mounts_network_attachments_and_image_digests", "all_business_writers_and_maintenance_locks", "database_identity_permissions_and_capacity", "network_connectivity_not_validated_by_this_metadata_report", "computer_held_backup_and_restore_materials", "business_and_external_projection_reconciliation", "production_scale_recovery_and_failure_drills"],
})
print("OK: reported container host ports and network internal boundaries match the selected Compose configuration")
print("OK: runtime evidence contains allowlisted metadata; target environment acceptance remains incomplete")
PY

if [[ "${CHECK_CONNECTIVITY}" != "true" ]]; then
  log "read-only runtime metadata checks complete; no diagnostic containers were started"
  exit 0
fi

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
run_diagnostic "api_ingress network reaches Go API through api-ingress alias" "${api_ingress_network_name}" "api-ingress" 9000 true
run_diagnostic "app network cannot reach PostgreSQL from storefront side" "${app_network_name}" "db" 5432 false

log "runtime connectivity checks complete; evidence written to ${REPORT_DIR}"

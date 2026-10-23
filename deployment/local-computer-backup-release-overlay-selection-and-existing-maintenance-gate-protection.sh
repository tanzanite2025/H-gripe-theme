#!/usr/bin/env bash

# Shared by deploy.sh and the release verifier. Only select Compose inputs and inspect
# existing containers/directories; they never start the maintenance executor.
append_redis_sentinel_overlay_for_explicit_release_mode() {
  local repository_directory="$1"
  local redis_mode="$2"
  local -n redis_compose_arguments="$3"
  case "${redis_mode,,}" in
    ""|standalone|single) return 0 ;;
    sentinel|failover)
      local overlay_file="${repository_directory}/deployment/redis-sentinel.compose.yml"
      if [[ ! -f "${overlay_file}" ]]; then
        echo "ERR: selected release does not include the Redis Sentinel overlay." >&2
        return 1
      fi
      redis_compose_arguments+=(-f "${overlay_file}")
      ;;
    *)
      echo "ERR: REDIS_MODE must be standalone or sentinel." >&2
      return 1
      ;;
  esac
}

append_local_computer_backup_overlay_for_explicit_release_setting() {
  local repository_directory="$1"
  local enabled="$2"
  local -n compose_arguments="$3"
  case "${enabled}" in
    false) return 0 ;;
    true)
      local overlay_file="${repository_directory}/compose.local-computer-backup-maintenance.yml"
      if [[ ! -f "${overlay_file}" ]]; then
        echo "ERR: selected release does not include the local computer backup maintenance overlay." >&2
        return 1
      fi
      compose_arguments+=(-f "${overlay_file}")
      ;;
    *)
      echo "ERR: LOCAL_BACKUP_MAINTENANCE_ENABLED must be true or false." >&2
      return 1
      ;;
  esac
}

verify_existing_local_computer_backup_gate_before_application_release() {
  local project_name="$1"
  local enabled="$2"
  local selected_gate_directory="$3"
  local selected_gate_identity=""
  local container_ids container_id service_name previous_gate_directory previous_gate_identity

  if [[ -z "${project_name}" || ! "${project_name}" =~ ^[a-z0-9][a-z0-9_-]*$ ]]; then
    echo "ERR: a resolved Compose project identity is required for backup gate inspection." >&2
    return 1
  fi
  if [[ "${enabled}" == "true" ]]; then
    if [[ "${selected_gate_directory}" != /* || ! -d "${selected_gate_directory}" ]]; then
      echo "ERR: the selected backup maintenance gate must already exist as an absolute host directory." >&2
      return 1
    fi
    selected_gate_identity="$(cd -- "${selected_gate_directory}" && pwd -P)" || return 1
    if [[ "${selected_gate_identity}" == "/" || ! -r "${selected_gate_directory}" || ! -x "${selected_gate_directory}" ]]; then
      echo "ERR: the deployment user must be able to inspect the dedicated backup maintenance gate." >&2
      return 1
    fi
    if [[ -e "${selected_gate_directory}/application-maintenance.json" || -L "${selected_gate_directory}/application-maintenance.json" ]]; then
      echo "ERR: an application backup/recovery fence is present; complete or reconcile maintenance before releasing application services." >&2
      return 1
    fi
  elif [[ "${enabled}" != "false" || "${selected_gate_directory}" != "-" ]]; then
    echo "ERR: local computer backup release selection and the resolved application gate disagree." >&2
    return 1
  fi

  # Include stopped containers: they can still retain the old recovery gate.
  container_ids="$(docker ps --all --quiet --filter "label=com.docker.compose.project=${project_name}")" || return 1
  while IFS= read -r container_id; do
    [[ -n "${container_id}" ]] || continue
    service_name="$(docker inspect --format '{{index .Config.Labels "com.docker.compose.service"}}' "${container_id}")" || return 1
    previous_gate_directory="$(docker inspect --format '{{range .Mounts}}{{if eq .Destination "/app/local-backup-maintenance-gate"}}{{.Source}}{{end}}{{end}}' "${container_id}")" || return 1
    if [[ "${service_name}" == "local-backup-maintenance" || -n "${previous_gate_directory}" ]]; then
      if [[ "${enabled}" != "true" ]]; then
        echo "ERR: existing project containers use local computer backup maintenance; refusing a release that omits its overlay." >&2
        return 1
      fi
      if [[ -z "${previous_gate_directory}" || ! -d "${previous_gate_directory}" ]]; then
        echo "ERR: an existing backup maintenance gate cannot be inspected; refusing to replace its recovery boundary." >&2
        return 1
      fi
      previous_gate_identity="$(cd -- "${previous_gate_directory}" && pwd -P)" || return 1
      if [[ "${previous_gate_identity}" != "${selected_gate_identity}" ]]; then
        echo "ERR: the selected release changes an existing backup maintenance gate; reconcile the recovery boundary first." >&2
        return 1
      fi
    fi
  done <<< "${container_ids}"
}

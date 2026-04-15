#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEFAULT_CONFIG="${ROOT_DIR}/etc/config.json"
ARGS=("$@")

has_config_flag=false
for arg in "${ARGS[@]}"; do
  if [[ "${arg}" == "--config" || "${arg}" == "--config-file" ]]; then
    has_config_flag=true
    break
  fi
done

if [[ "${has_config_flag}" == false ]]; then
  # go run uses a temporary executable path, so pass the real repo config
  # explicitly unless the caller already selected a config file.
  ARGS=("--config-file" "${DEFAULT_CONFIG}" "${ARGS[@]}")
fi

cd "${ROOT_DIR}/src"
GOTOOLCHAIN=local GOCACHE=/tmp/dynamic-ip-updater-gocache go run . "${ARGS[@]}"

#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_PATH="${ROOT_DIR}/bin/diu"

mkdir -p "${ROOT_DIR}/bin"

cd "${ROOT_DIR}/src"
GOTOOLCHAIN=local GOCACHE=/tmp/dynamic-ip-updater-gocache go build -o "${OUTPUT_PATH}" .

echo "Built ${OUTPUT_PATH}"

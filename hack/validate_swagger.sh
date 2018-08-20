#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

readonly REPO_BASE="$(cd .. && pwd -P)"

source "${REPO_BASE}/hack/codegen/swagger.sh"

main() {
  local has_installed

  has_installed="$(swagger::check_version)"
  if [[ "${has_installed}" = "false" ]]; then
    echo >&2 "swagger-${SWAGGER_VERSION} should be installed."
    exit 1
  fi

  swagger validate "${REPO_BASE}/apis/swagger.yml"
}

main

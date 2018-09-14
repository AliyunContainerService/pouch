#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"
source "../codegen/swagger.sh"

# ci_related::install_gometalinter installs gometalinter for linters.
ci_related::install_gometalinter() {
  local has_installed

  has_installed="$(command -v gometalinter || echo false)"
  if [[ "${has_installed}" != "false" ]]; then
    echo "gometalinter has been installed."
    return
  fi

  go get -u github.com/alecthomas/gometalinter
  gometalinter --install > /dev/null
}

# ci_related::install_gocovmerge installs gocovmerge for coverage combine.
ci_related::install_gocovmerge() {
  local has_installed

  has_installed="$(command -v gocovmerge || echo false)"
  if [[ "${has_installed}" != "false" ]]; then
    echo "gocovmerge has been installed."
    return
  fi

  go get -u github.com/wadey/gocovmerge
}

# ci_related::install_swagger installs swagger binary.
ci_related::install_swagger() {
  local has_installed

  has_installed="$(swagger::check_version)"
  if [[ "${has_installed}" = "true" ]]; then
    echo "swagger-${SWAGGER_VERSION} has been installed."
    exit 0
  fi

  swagger::install
  command -v swagger
}

main() {
  echo "install CI related tools..."
  ci_related::install_gometalinter
  ci_related::install_gocovmerge
  ci_related::install_swagger
  echo
}

main

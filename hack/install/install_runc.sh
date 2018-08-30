#!/usr/bin/env bash

set -euo pipefail

readonly RUNC_VERSION="1.0.0-rc4-2"

# runc::check_version checks the command and the version.
runc::check_version() {
  local has_installed version

  has_installed="$(command -v runc || echo false)"
  if [[ "${has_installed}" = "false" ]]; then
    echo false
    exit 0
  fi

  version="$(runc -v | head -1 | cut -d " " -f 3)"
  if [[ "${RUNC_VERSION}" != "${version}" ]]; then

    echo false
    exit 0
  fi

  echo true
}

# runc::install downloads the binary from release url.
runc::install() {
  local url

  url="https://github.com/alibaba/runc/releases/download"
  url="${url}/v${RUNC_VERSION}/runc.amd64"

  wget --quiet "${url}" -P /usr/local/bin
  chmod +x /usr/local/bin/runc.amd64
  mv /usr/local/bin/runc.amd64 /usr/local/bin/runc
}

main() {
  local has_installed

  has_installed="$(runc::check_version)"
  if [[ "${has_installed}" = "true" ]]; then
    echo "runc-${RUNC_VERSION} has been installed."
    exit 0
  fi

  echo ">>>> install runc-${RUNC_VERSION} <<<<"

  runc::install

  # final check
  command -v runc > /dev/null

  echo
}

main

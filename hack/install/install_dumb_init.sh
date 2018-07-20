#!/usr/bin/env bash

set -euo pipefail

readonly DUMB_INIT_VERSION="1.2.1"

# dumb_init::check_version checks the command and the version.
dumb_init::check_version() {
  local has_installed version

  has_installed="$(command -v dumb-init || echo false)"
  if [[ "${has_installed}" = "false" ]]; then
    echo false
    exit 0
  fi

  version="$(dumb-init -V 2>&1 | cut -d " " -f 2 | cut -c 2-)"
  if [[ "${DUMB_INIT_VERSION}" != "${version}" ]]; then
    echo false
    exit 0
  fi

  echo true
}

# dumb_init::install downloads the binary from release url.
dumb_init::install() {
  local url target

  target="/tmp/dumb-init"

  url="https://github.com/Yelp/dumb-init/releases/download"
  url="${url}/v${DUMB_INIT_VERSION}/dumb-init_${DUMB_INIT_VERSION}_amd64"

  wget --quiet -O "${target}" "${url}"
  mv "${target}" /usr/bin/
  chmod +x /usr/bin/dumb-init
}

main() {
  local has_installed

  has_installed="$(dumb_init::check_version)"
  if [[ "${has_installed}" = "true" ]]; then
    echo "dumb-init-${DUMB_INIT_VERSION} has been installed."
    exit 0
  fi

  echo ">>>> install dumb-init-${DUMB_INIT_VERSION} <<<<"

  dumb_init::install

  # final check
  command -v dumb-init > /dev/null

  echo
}

main

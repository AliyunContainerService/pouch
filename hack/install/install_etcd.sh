#!/usr/bin/env bash

set -euo pipefail

ETCD_VERSION=v3.3.5
ARCH=amd64

# keep the first one only
GOPATH="${GOPATH%%:*}"

# add bin folder into PATH.
export PATH="${GOPATH}/bin:${PATH}"

# etcd::check_version checks the command and the version.
etcd::check_version() {
  local has_installed

  has_installed="$(command -v etcd || echo false)"
  if [[ "${has_installed}" = "false" ]]; then
    echo false
    exit 0
  fi

  echo true
}

# etcd::install downloads the package and build.
etcd::install() {
  wget --quiet "https://github.com/coreos/etcd/releases/download/${ETCD_VERSION}/etcd-${ETCD_VERSION}-linux-${ARCH}.tar.gz"
  tar -xf "etcd-${ETCD_VERSION}-linux-${ARCH}.tar.gz" -C "/usr/local"
  rm etcd-${ETCD_VERSION}-linux-${ARCH}.tar.gz
  export PATH="/usr/local/etcd-${ETCD_VERSION}-linux-${ARCH}:${PATH}"
}

main() {
  local has_installed

  has_installed="$(etcd::check_version)"
  if [[ "${has_installed}" = "true" ]]; then
    echo "etcd-${ETCD_VERSION} has been installed."
    exit 0
  fi

  echo ">>>> install etcd-${ETCD_VERSION} <<<<"
  etcd::install

  command -v etcd > /dev/null
}

main "$@"

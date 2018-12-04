#!/usr/bin/env bash

set -euo pipefail

readonly CONTAINERD_VERSION="v1.2.0"

# containerd::check_version checks the command and the version.
containerd::check_version() {
  local has_installed version

  has_installed="$(command -v containerd || echo false)"
  if [[ "${has_installed}" = "false" ]]; then
    echo false
    exit 0
  fi

  version="$(containerd -v | cut -d " " -f 3 | cut -c 2-)"
  if [[ "${CONTAINERD_VERSION}" != "${version}" ]]; then
    echo false
    exit 0
  fi

  echo true
}

# containerd::install downloads the code and builds it.
containerd::install() {
  local tmpdir pkgpath

  # create gopath
  tmpdir="$(mktemp -d /tmp/containerd-install-XXXXXX)"
  trap 'rm -rf /tmp/containerd-install-*' EXIT

  pkgpath="$tmpdir/src/github.com/containerd/containerd"
  mkdir -p "${pkgpath}"

  git clone https://github.com/containerd/containerd "${pkgpath}"
  cd "${pkgpath}"

  git checkout "${CONTAINERD_VERSION}"
  GOPATH=$tmpdir make BUILDTAGS=no_cri # build without cri plugin
  make install
}

main() {
  local has_installed

  has_installed="$(containerd::check_version)"
  if [[ "${has_installed}" = "true" ]]; then
    echo "containerd-${CONTAINERD_VERSION} has been installed."
    exit 0
  fi

  echo ">>>> install containerd-${CONTAINERD_VERSION} <<<<"

  containerd::install

  # final check
  command -v containerd > /dev/null

  echo
}

main

#!/usr/bin/env bash

set -euo pipefail

readonly CONTAINERD_VERSION="1.0.3"

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

# containerd::install downloads the binary from release url.
containerd::install() {
  local url target tmpdir

  target="containerd-${CONTAINERD_VERSION}.linux-amd64.tar.gz"
  url="https://github.com/containerd/containerd/releases/download"
  url="${url}/v${CONTAINERD_VERSION}/${target}"

  tmpdir="$(mktemp -d /tmp/containerd-install-XXXXXX)"
  trap 'rm -rf /tmp/containerd-install-*' EXIT

  wget --quiet "${url}" -P "${tmpdir}"
  tar xf "${tmpdir}/${target}" -C "${tmpdir}"
  cp -f "${tmpdir}"/bin/* /usr/local/bin/
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

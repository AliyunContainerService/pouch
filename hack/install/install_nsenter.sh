#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"
source "./check.sh"

readonly NSENTER_VERSION="2.24.1"
readonly cmd="nsenter"

# nsenter::check_version checks the command and the version.
nsenter::check_version() {
  local has_installed version

  has_installed="$(command -v nsenter || echo false)"
  if [[ "${has_installed}" = "false" ]]; then
    echo false
    exit 0
  fi

  version="$(nsenter -V | cut -d " " -f 4)"
  if [[ "${NSENTER_VERSION}" != "${version}" ]]; then
    echo false
    exit 0
  fi

  echo true
}

# nsenter::ubuntu::install_dependencies will use apt-get to install dependencies.
nsenter::ubuntu::install_dependencies() {
  apt-get install -y -q \
    libncurses5-dev \
    libslang2-dev \
    gettext \
    zlib1g-dev \
    libselinux1-dev \
    debhelper \
    lsb-release \
    pkg-config \
    po-debconf \
    autoconf \
    automake \
    autopoint \
    libtool
}

# nsenter::ubuntu::install will install nsenter.
nsenter::ubuntu::install() {
  local url target tmpdir

  target="util-linux-${NSENTER_VERSION}.tar.gz"
  url="https://www.kernel.org/pub/linux/utils/util-linux/v2.24"
  url="${url}/${target}"

  tmpdir="$(mktemp -d /tmp/nsenter-install-XXXXXX)"
  trap 'rm -rf /tmp/nsenter-install-*' EXIT

  wget --quiet "${url}" -P "${tmpdir}"
  tar xf "${tmpdir}/${target}" -C "${tmpdir}"

  cd "${tmpdir}/util-linux-${NSENTER_VERSION}"
  ./autogen.sh
  autoreconf -vfi
  ./configure
  make

  cp "${cmd}" /usr/local/bin
}

# nsenter::centos::install will install nsenter.
nsenter::centos::install() {
  yum install -y -q util-linux
}

main() {
  local os_dist has_installed

  has_installed="$(nsenter::check_version)"
  if [[ "${has_installed}" == "true" ]]; then
    echo "${cmd}-${NSENTER_VERSION} has been installed!"
    exit 0
  fi

  echo ">>>> install ${cmd}-${NSENTER_VERSION} <<<<"

  os_dist="$(detect_os)"
  if [[ "${os_dist}" = "Ubuntu" ]]; then
    nsenter::ubuntu::install_dependencies > /dev/null
    nsenter::ubuntu::install > /dev/null
  else
    nsenter::centos::install > /dev/null
  fi

  # final check
  command -v "${cmd}" > /dev/null

  echo
}

main

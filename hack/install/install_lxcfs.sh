#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"
source "./check.sh"

# lxcfs::centos::install_dependencies will use yum to install dependencies.
lxcfs::centos::install_dependencies() {
  local arch
  arch="$(uname -p)"

  yum install -y -q \
    autobook \
    autoconf-doc \
    autoconf2.13 \
    autotools-dev \
    autoconf-archive \
    git \
    gnu-standards \
    libtool \
    make \
    m4 \
    "fuse-devel.${arch}" \
    "pam-devel.${arch}" \
    "fuse.${arch}"
}

# lxcfs::centos::pull_build will pull stable-2.0 to build lxcfs.
lxcfs::centos::pull_build() {
  local tmpdir branch

  tmpdir="$(mktemp -d /tmp/lxcfs-build-XXXXXX)"
  branch="stable-2.0"

  trap 'rm -rf /tmp/lxcfs-build-*' EXIT

  git clone -b "${branch}" https://github.com/lxc/lxcfs.git "${tmpdir}/lxcfs"
  cd "${tmpdir}/lxcfs"

  ./bootstrap.sh
  ./configure
  make install

  ln -s /usr/local/bin/lxcfs /usr/bin/lxcfs
  mkdir -p /var/lib/lxcfs
}

# lxcfs::ubuntu::install will use repository manager to install lxcfs.
#
# FIXME: It is 3.0.0 version which different with stable-2.0 in centos!!!!!!!
lxcfs::ubuntu::install() {
  add-apt-repository -y ppa:ubuntu-lxc/lxcfs-stable
  apt-get update -q -y
  apt-get install -y -q lxcfs
}

main() {
  local os_dist has_installed

  has_installed="$(command -v lxcfs || echo false)"
  if [[ "${has_installed}" != "false" ]]; then
    echo "lxcfs has been installed!"
    exit 0
  fi

  echo ">>>> install lxcfs <<<<"

  os_dist="$(detect_os)"
  if [[ "${os_dist}" = "Ubuntu" ]]; then
    lxcfs::ubuntu::install > /dev/null
  else
    lxcfs::centos::install_dependencies > /dev/null
    lxcfs::centos::pull_build > /dev/null
  fi

  # final check
  command -v lxcfs > /dev/null

  echo
}

main

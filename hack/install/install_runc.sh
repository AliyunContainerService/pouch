#!/usr/bin/env bash

set -euo pipefail

# keep the first one only
GOPATH="${GOPATH%%:*}"

cd "$(dirname "${BASH_SOURCE[0]}")"
source "./check.sh"

# runc::ubuntu::install_dependencies() install dependencies
# on ubuntu machine for make runc
runc::ubuntu::install_dependencies() {
  sudo apt-get install -y libseccomp-dev/trusty-backports
}

# runc::centos::install_dependencies() install dependencies 
# on centos machine for make runc
runc::centos::install_dependencies() {
  sudo yum install libseccomp-dev
}

# runc::check_install checks the command and the version.
runc::check_install() {
  local has_installed

  has_installed="$(command -v runc || echo false)"
  if [[ "${has_installed}" = "false" ]]; then
    echo false
    return
  fi

  echo true
}

# runc::install downloads the binary from release url.
runc::install() {
  local gopath

  gopath="${GOPATH}/src/github.com/opencontainers/runc"
  git clone -b develop https://github.com/alibaba/runc.git "${gopath}"

  cd "${gopath}"
  make
  make install
}

main() {
  local has_installed

  has_installed="$(runc::check_install)"
  if [[ "${has_installed}" = "true" ]]; then
    echo "runc has been installed."
    exit 0
  fi

  echo ">>>> install runc <<<<"

  os_dist="$(detect_os)"
  if [[ "${os_dist}" = "Ubuntu" ]]; then
    runc::ubuntu::install_dependencies
  else
    runc::centos::install_dependencies
  fi
  runc::install

  # final check
  command -v runc > /dev/null

  echo

  # test version runc
  runc -v
}

main

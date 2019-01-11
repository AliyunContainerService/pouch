#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"
source "./check.sh"
source "./config.sh"

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
  apt-get install -y -q wget
}

# nsenter::ubuntu::install will install nsenter.
# TODO: change to get binary from aliyun oss storage.
nsenter::ubuntu::install() {
  wget "https://${OSS_BUCKET}.${OSS_ENDPOINT}/pouch-test/ubuntu/nsenter-2.24.1" \
    -O /usr/local/bin/nsenter
  chmod +x /usr/local/bin/nsenter
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

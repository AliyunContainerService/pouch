#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"
source "./check.sh"

# criu::ubuntu::install will install criu from source.
criu::ubuntu::install() {
  apt-get install -y -q criu
}

main() {
  local os_dist has_installed

  has_installed="$(command -v criu || echo false)"
  if [[ "${has_installed}" != "false" ]]; then
    echo "criu has been installed!"
    exit 0
  fi

  echo ">>>> install criu <<<<"

  os_dist="$(detect_os)"
  if [[ "${os_dist}" = "Ubuntu" ]]; then
    criu::ubuntu::install > /dev/null
  else
    echo "will support redhat soon"
    exit 0
  fi

  # final check
  command -v criu > /dev/null

  echo
}

main

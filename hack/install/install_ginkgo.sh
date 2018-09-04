#!/usr/bin/env bash

set -euo pipefail

# keep the first one only
GOPATH="${GOPATH%%:*}"

# add bin folder into PATH.
export PATH="${GOPATH}/bin:${PATH}"


# ginkgo::install_ginkgo installs ginkgo if missing.
ginkgo::install_ginkgo() {
  local has_installed pkg

  pkg="github.com/onsi/ginkgo/ginkgo"
  has_installed="$(command -v ginkgo || echo false)"
  if [[ "${has_installed}" = "false" ]]; then
    go get -u "${pkg}"
  fi

  command -v ginkgo > /dev/null
}

main() {
  ginkgo::install_ginkgo
}

main "$@"

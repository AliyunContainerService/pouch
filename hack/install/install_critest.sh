#!/usr/bin/env bash

set -euo pipefail

CRITEST_BRANCH_DEFAULT=master

# keep the first one only
GOPATH="${GOPATH%%:*}"

# add bin folder into PATH.
export PATH="${GOPATH}/bin:${PATH}"

# critest::check_version checks the command and the version.
critest::check_version() {
  local has_installed version

  has_installed="$(command -v critest || echo false)"
  if [[ "${has_installed}" = "false" ]]; then
    echo false
    exit 0
  fi

  version="$(critest -version | head -n 1 | cut -d " " -f 3)"
  if [[ "${CRITEST_VERSION}" != "${version}" ]]; then
    echo false
    exit 0
  fi

  echo true
}

# critest::install downloads the package and build.
critest::install() {
  local workdir pkg CRITOOLS_REPO

  pkg="github.com/kubernetes-sigs/cri-tools"
  CRITOOLS_REPO="github.com/alibaba/cri-tools"
  workdir="${GOPATH}/src/${pkg}"

  if [ ! -d "${workdir}" ]; then
    mkdir -p "${workdir}"
    cd "${workdir}"
    git clone https://${CRITOOLS_REPO} .
  fi

  cd "${workdir}"
  git fetch --all
  git checkout "${CRITEST_BRANCH_DEFAULT}"
  make
  cd -
}

# critest::install_ginkgo installs ginkgo if missing.
critest::install_ginkgo() {
  hack/install/install_ginkgo.sh
}

# critest::install_socat installs socat if missing.
critest::install_socat() {
  sudo apt-get install -y socat
}

main() {
  critest::install_ginkgo
  critest::install_socat

  local has_installed

  CRITEST_VERSION="1.0.0-beta.0"
  has_installed="$(critest::check_version)"
  if [[ "${has_installed}" = "true" ]]; then
    echo "critest-${CRITEST_VERSION} has been installed."
    exit 0
  fi

  echo ">>>> install critest-${CRITEST_VERSION} <<<<"
  critest::install

  command -v critest > /dev/null
}

main "$@"

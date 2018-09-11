#!/usr/bin/env bash

set -euo pipefail

CRITEST_VERSION=1.0.0-beta.0
CRITEST_BRANCH=tools-dev

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
  local workdir pkg cri_runtime CRITOOLS_REPO
  cri_runtime=$1

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
  if [[ "${cri_runtime}" == "v1alpha1" ]]; then
      git checkout "v${CRITEST_VERSION}"
  else
      git checkout "${CRITEST_BRANCH}"
  fi
  make
  cd -
}

# critest::install_ginkgo installs ginkgo if missing.
critest::install_ginkgo() {
  hack/install/install_ginkgo.sh
}

main() {
  critest::install_ginkgo

  local cri_runtime has_installed
  cri_runtime=$1

  if [[ "${cri_runtime}" == "v1alpha1" ]]; then
      CRITEST_VERSION="1.0.0-alpha.0"
  else
      CRITEST_VERSION="1.0.0-beta.0"
  fi

  has_installed="$(critest::check_version)"
  if [[ "${has_installed}" = "true" ]]; then
    echo "critest-${CRITEST_VERSION} has been installed."
    exit 0
  fi

  echo ">>>> install critest-${CRITEST_VERSION} <<<<"
  critest::install "${cri_runtime}"

  command -v critest > /dev/null
}

main "$@"

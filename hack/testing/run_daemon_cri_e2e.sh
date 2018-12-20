#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"
source utils.sh

cd ../../
readonly REPO_BASE="$(pwd -P)"

# keep the first one only
GOPATH="${GOPATH%%:*}"

# add bin folder into PATH so that pouch-e2e is available.
export PATH="${REPO_BASE}/bin:${PATH}"

# add bin folder into PATH.
export PATH="${GOPATH}/bin:${PATH}"

# CRI_SKIP skips the test to skip.
DEFAULT_SKIP="\[Flaky\]|\[Slow\]|\[Serial\]"
DEFAULT_SKIP+="|querying\s\/stats\/summary"
DEFAULT_SKIP+="|should execute prestop exec hook properly"
DEFAULT_SKIP+="|should execute poststart exec hook properly"
DEFAULT_SKIP+="|should function for intra-pod communication: http*"
DEFAULT_SKIP+="|should function for intra-pod communication: udp*"
export SKIP=${SKIP:-${DEFAULT_SKIP}}

# FOCUS focuses the test to run.
DEFAULT_FOCUS="\[NodeConformance\]"
export FOCUS=${FOCUS:-${DEFAULT_FOCUS}}

POUCH_SOCK="/var/run/pouchcri.sock"

# tmplog_dir stores the background job log data
tmplog_dir="$(mktemp -d /tmp/e2e-daemon-cri-testing-XXXXX)"
pouchd_log="${tmplog_dir}/pouchd.log"
local_persist_log="${tmplog_dir}/local_persist.log"
trap 'rm -rf /tmp/e2e-daemon-cri-testing-*' EXIT

# integration::install_cni installs cni plugins.
integration::install_cni() {
   hack/install/install_cni.sh
}

# integration::install_ginkgo installs ginkgo.
integration::install_ginkgo() {
  hack/install/install_ginkgo.sh
}

# integration::install_etcd installs etcd.
integration::install_etcd() {
  hack/install/install_etcd.sh
}

# integration::run_daemon_cri_test_e2e_cases runs CRI e2e test cases.
integration::run_daemon_cri_test_e2e_cases() {
  local cri_runtime code KUBERNETES_VERSION
  cri_runtime=$1

  if [[ "${cri_runtime}" == "v1alpha1" ]]; then
    KUBERNETES_VERSION="release-1.9"
  else
    KUBERNETES_VERSION="release-1.12"
  fi

  KUBERNETES_REPO="github.com/kubernetes/kubernetes"
  KUBERNETES_PATH="${GOPATH}/src/k8s.io/kubernetes"
  if [ ! -d "${KUBERNETES_PATH}" ]; then
    mkdir -p "${KUBERNETES_PATH}"
    cd "${KUBERNETES_PATH}"
    git clone https://${KUBERNETES_REPO} .
  fi
  cd "${KUBERNETES_PATH}"
  git fetch --all
  git checkout ${KUBERNETES_VERSION}

  echo "start pouch daemon cri-${cri_runtime} e2e test..."
  set +e

  make test-e2e-node \
    RUNTIME=remote \
    CONTAINER_RUNTIME_ENDPOINT=unix://${POUCH_SOCK} \
    SKIP="${SKIP}" \
    FOCUS="${FOCUS}" \
    TEST_ARGS='--kubelet-flags="--cgroups-per-qos=true --cgroup-root=/"' \
    PARALLELISM=8

  code=$?

  integration::stop_local_persist
  integration::stop_pouchd
  set -e

  if [[ "${code}" != "0" ]]; then
    echo "failed to pass e2e cases!"
    echo "there is daemon logs..."
    cat "${pouchd_log}"
    exit ${code}
  fi

  # sleep for pouchd stop and got the coverage
  sleep 5
}

integration::run_cri_e2e_test(){
  local cri_runtime cmd flags coverage_profile
  cri_runtime=$1

  # daemon cri integration coverage profile
  coverage_profile="${REPO_BASE}/coverage/e2e_daemon_cri_${cri_runtime}_profile.out"
  rm -rf "${coverage_profile}"
  
  cmd="pouchd-integration"
  flags=" -test.coverprofile=${coverage_profile} DEVEL"
  flags="${flags} --enable-cri --cri-version ${cri_runtime} --sandbox-image=gcr.io/google_containers/pause-amd64:3.0"


  integration::stop_local_persist
  integration::run_local_persist_background "${local_persist_log}"
  integration::stop_pouchd
  integration::run_pouchd_background "${cmd}" "${flags}" "${pouchd_log}"

  set +e; integration::ping_pouchd; code=$?; set -e
  if [[ "${code}" != "0" ]]; then
    echo "there is daemon logs..."
    cat "${pouchd_log}"
    exit ${code}
  fi
  integration::run_daemon_cri_test_e2e_cases "${cri_runtime}"
}

main() {
  local cri_runtime
  cri_runtime=$1

  integration::install_cni
  integration::install_ginkgo
  integration::install_etcd
  integration::run_cri_e2e_test "${cri_runtime}"
}

main "$@"

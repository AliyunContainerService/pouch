#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"
source utils.sh

cd ../../
readonly REPO_BASE="$(pwd -P)"

# keep the first one only
GOPATH="${GOPATH%%:*}"

# add bin folder into PATH so that pouch-integration is available.
export PATH="${REPO_BASE}/bin:${PATH}"

# add bin folder into PATH.
export PATH="${GOPATH}/bin:${PATH}"

# CRI_SKIP skips the test to skip.
DEFAULT_CRI_SKIP="RunAsUserName|seccomp localhost"
DEFAULT_CRI_SKIP="${DEFAULT_CRI_SKIP}|should error on create with wrong options"
DEFAULT_CRI_SKIP="${DEFAULT_CRI_SKIP}|runtime should support reopening container log"
CRI_SKIP="${CRI_SKIP:-"${DEFAULT_CRI_SKIP}"}"

# CRI_FOCUS focuses the test to run.
# With the CRI manager completes its function, we may need to expand this field.
CRI_FOCUS=${CRI_FOCUS:-}

POUCH_SOCK="/var/run/pouchcri.sock"

# tmplog_dir stores the background job log data
tmplog_dir="$(mktemp -d /tmp/integration-daemon-cri-testing-XXXXX)"
pouchd_log="${tmplog_dir}/pouchd.log"
local_persist_log="${tmplog_dir}/local_persist.log"
trap 'rm -rf /tmp/integration-daemon-cri-testing-*' EXIT

# integration::install_critest installs test case.
integration::install_critest() {
  local cri_runtime
  cri_runtime=$1
  hack/install/install_critest.sh "${cri_runtime}"
}

# integration::install_cni installs cni plugins.
integration::install_cni() {
   hack/install/install_cni.sh
}


# integration::run_daemon_cri_test_cases runs CRI test cases.
integration::run_daemon_cri_test_cases() {
  local cri_runtime code
  cri_runtime=$1
  echo "start pouch daemon cri-${cri_runtime} integration test..."

  set +e
  if [[ "${cri_runtime}" == "v1alpha1" ]]; then
    critest --runtime-endpoint=${POUCH_SOCK} \
      --focus="${CRI_FOCUS}" --ginkgo-flags="--skip=\"${CRI_SKIP}\"" validation
  else
    critest --runtime-endpoint=${POUCH_SOCK} \
      --ginkgo.focus="${CRI_FOCUS}" --ginkgo.skip="${CRI_SKIP}"
  fi
  code=$?

  integration::stop_local_persist
  integration::stop_pouchd
  set -e

  if [[ "${code}" != "0" ]]; then
    echo "failed to pass integration cases!"
    echo "there is daemon logs..."
    cat "${pouchd_log}"
    exit ${code}
  fi

  # sleep for pouchd stop and got the coverage
  sleep 5
}

integration::run_cri_test(){
  local cri_runtime cmd flags coverage_profile
  cri_runtime=$1

  # daemon cri integration coverage profile
  coverage_profile="${REPO_BASE}/coverage/integration_daemon_cri_${cri_runtime}_profile.out"
  rm -rf "${coverage_profile}"
  
  cmd="pouchd-integration"
  flags=" -test.coverprofile=${coverage_profile} DEVEL"
  flags="${flags} --enable-cri --cri-version ${cri_runtime} --sandbox-image=gcr.io/google_containers/pause-amd64:3.0"

  integration::install_critest "${cri_runtime}"

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
  integration::run_daemon_cri_test_cases "${cri_runtime}"
}

main() {
  local cri_runtime
  cri_runtime=$1

  integration::install_cni
  integration::run_cri_test "${cri_runtime}"
}

main "$@"

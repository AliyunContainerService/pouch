#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"
source utils.sh

readonly REPO_BASE="$(cd ../../ && pwd -P)"

# add bin folder into PATH so that pouch-integration is available.
export PATH="${REPO_BASE}/bin:${PATH}"

# tmplog_dir stores the background job log data
tmplog_dir="$(mktemp -d /tmp/integration-testing-XXXXX)"
pouchd_log="${tmplog_dir}/pouchd.log"
local_persist_log="${tmplog_dir}/local_persist.log"
trap 'rm -rf /tmp/integration-testing-*' EXIT

# daemon integration coverage profile
coverage_profile="${REPO_BASE}/coverage/integration_daemon_profile.out"
rm -rf "${coverage_profile}"

# integration::run_daemon_test_cases starts cases.
integration::run_daemon_test_cases() {
  echo "start pouch daemon integration test..."
  local code=0
  local job_id=$1

  cp -rf "${REPO_BASE}/test/tls" /tmp/

  set +e
  pushd "${REPO_BASE}/test"
  local testcases
  testcases=$(cat "${REPO_BASE}/test/testcase.list.${job_id}")
  for one in ${testcases}; do
    go test -check.v -check.f "${one}"
    ret=$?
    if [[ ${ret} -ne 0 ]]; then
      code=${ret}
    fi
  done

  integration::stop_local_persist
  integration::stop_pouchd
  set -e

  if [[ "${code}" != "0" ]]; then
    echo "failed to pass integration cases!"
    echo "there is daemon logs...."
    cat "${pouchd_log}"
    exit ${code}
  fi

  # sleep for pouchd stop and got the coverage
  sleep 5
}

main() {
  local cmd flags
  local job_id=$1
  cmd="pouchd-integration"
  flags=" -test.coverprofile=${coverage_profile} DEVEL"
  flags="${flags} --debug --enable-lxcfs --add-runtime runv=runv"

  integration::stop_local_persist
  integration::run_local_persist_background "${local_persist_log}"

  integration::stop_mount_lxcfs
  integration::run_mount_lxcfs_background

  integration::stop_pouchd
  integration::run_pouchd_background "${cmd}" "${flags}" "${pouchd_log}"

  set +e; integration::ping_pouchd; code=$?; set -e
  if [[ "${code}" != "0" ]]; then
    echo "there is daemon logs..."
    cat "${pouchd_log}"
    exit ${code}
  fi
  integration::run_daemon_test_cases "${job_id}"
}

main "$@"

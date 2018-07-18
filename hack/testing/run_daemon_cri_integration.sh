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

# daemon cri integration coverage profile
coverage_profile="${REPO_BASE}/coverage/integration_daemon_cri_profile.out"
rm -rf "${coverage_profile}"


# integration::install_critest installs test case.
integration::install_critest() {
  hack/install/install_critest.sh
}

# integration::install_cni installs cni plugins.
integration::install_cni() {
  echo "install cni..."

  local workdir pkg

  # for multiple GOPATHs, keep the first one only
  pkg="github.com/containernetworking/plugins"
  workdir="${GOPATH}/src/${pkg}"

  # downloads github.com/containernetworking/plugins
  go get -u -d "${pkg}"/...

  # build and copy into /opt/cni/bin
  "${workdir}"/build.sh
  mkdir -p /etc/cni/net.d /opt/cni/bin
  cp "${workdir}"/bin/* /opt/cni/bin

  # setup the config
  sh -c 'cat >/etc/cni/net.d/10-mynet.conflist <<-EOF
{
    "cniVersion": "0.3.1",
    "name": "mynet",
    "plugins": [
        {
            "type": "bridge",
            "bridge": "cni0",
            "isGateway": true,
            "ipMasq": true,
            "ipam": {
                "type": "host-local",
                "subnet": "10.30.0.0/16",
                "routes": [
                    { "dst": "0.0.0.0/0"   }
                ]
            }
        },
        {
            "type": "portmap",
            "capabilities": {"portMappings": true},
            "snat": true
        }
    ]
}
EOF'

  sh -c 'cat >/etc/cni/net.d/99-loopback.conf <<-EOF
{
    "cniVersion": "0.3.1",
    "type": "loopback"
}
EOF'

  echo
}


# integration::run_daemon_cri_test_cases runs CRI test cases.
integration::run_daemon_cri_test_cases() {
  echo "start pouch daemon cri integration test..."
  local code

  set +e
  critest --runtime-endpoint=${POUCH_SOCK} \
    --ginkgo.focus="${CRI_FOCUS}" --ginkgo.skip="${CRI_SKIP}"
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

main() {
  local cmd flags
  cmd="pouchd-integration"
  flags=" -test.coverprofile=${coverage_profile} DEVEL"
  flags="${flags} --enable-cri --sandbox-image=gcr.io/google_containers/pause-amd64:3.0"

  integration::install_cni
  integration::install_critest

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
  integration::run_daemon_cri_test_cases
}

main "$@"

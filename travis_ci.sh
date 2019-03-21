#!/bin/bash

set -e # fast fail

# action for travis ci
Action=$1
# config the number of integration test job.
JOBS=$2
# config the id of job
JOB_ID=$3

CASE_PREFIX="testcase.list"
LOCAL_CASE_FILE="${TRAVIS_BUILD_DIR}/test/${CASE_PREFIX}"

source "${TRAVIS_BUILD_DIR}/hack/install/config.sh"

prepare_testcase() {
  # get the number of all test case
  # test case must write with this format:
  # func (XXX) TestXXX(c *check.C) {
  grep -h -E "^func\ \(.*\)\ Test.*\(c\ \*check\.C\)\ \{" ./test -r \
    | awk '{print $3,$4}' | awk -F \( '{print $1}' \
    | sed 's/\*//' | sed 's/) /./' > "${LOCAL_CASE_FILE}"

  # make test case list file
  local sum nums

  # shellcheck disable=SC2002
  sum=$(cat "${LOCAL_CASE_FILE}" | wc -l)
  nums=$((sum / JOBS + 1))

  rm -rf "${LOCAL_CASE_FILE}.*"
  split -l ${nums} -d --suffix-length=1 "${LOCAL_CASE_FILE}" "${LOCAL_CASE_FILE}."
}

run_pre_test() {
  prepare_testcase
}

run_unittest() {
  sudo env "PATH=$PATH" hack/install/install_ci_related.sh
  sudo env "PATH=$PATH" make unit-test
  make coverage

  bash <(curl -s https://codecov.io/bash) -cF unittest -y .codecov.yml
}

run_integration_test() {
  local job_id=$1

  check_port

  make build
  sudo env "PATH=$PATH" make install

  sudo env "PATH=$PATH" make download-dependencies
  sudo env "PATH=$PATH" "INTEGRATION_FLAGS=${job_id}" make integration-test
  make coverage

  bash <(curl -s https://codecov.io/bash) -cF "integration_test_${job_id}" -y .codecov.yml
}

run_criv1alpha2_test() {
  check_port

  make build
  TEST_FLAGS="" BUILDTAGS="selinux seccomp apparmor" make build-daemon-integration
  sudo env "PATH=$PATH" make install

  sudo env "PATH=$PATH" make download-dependencies
  sudo env "PATH=$PATH" make cri-v1alpha2-test
  make coverage

  bash <(curl -s https://codecov.io/bash) -cF criv1alpha2_test -y .codecov.yml
}

run_node_e2e_test() {
  check_port

  make build
  TEST_FLAGS="" make build-daemon-integration
  sudo env "PATH=$PATH" make install

  sudo env "PATH=$PATH" make download-dependencies
  sudo env "PATH=$PATH" make cri-e2e-test
  make coverage

  bash <(curl -s https://codecov.io/bash) -cF node_e2e_test -y .codecov.yml
}

check_port() {
	# docker-containerd will use port 10010 in docker.service
	sudo netstat -npl | grep -q "10010"
	if [[ $? -ne 0 ]]; then
		return
	fi

	sudo systemctl stop docker
}

install_osscmd() {
  sudo wget "http://gosspublic.alicdn.com/ossutil/1.4.1/ossutil64" \
      -O /usr/local/bin/ossutil
  sudo chmod +x /usr/local/bin/ossutil
}

oss() {
  local action=$1
  local src=$2
  local dst=$3

  command -v ossutil || install_osscmd

  local osscmd="ossutil -e ${OSS_ENDPOINT} \
    -i ${OSS_ACCESS_KEY_ID} \
    -k ${OSS_ACCESS_KEY_SECRET}"

  case ${action} in
    ls)
      ${osscmd} ls "${src}"
    ;;
    cp)
      ${osscmd} cp -r -f "${src}" "${dst}"
    ;;
    rm)
      ${osscmd} rm -rf "${src}"
    ;;
    *)
      echo "Unsupport command"
    ;;
  esac
}

main () {
  case ${Action} in
    pretest)
      echo "pre-test"
      run_pre_test
    ;;
    unittest)
      echo "run unit test"
      run_unittest
    ;;
    integrationtest)
      echo "run integration test"
      run_pre_test
      run_integration_test "${JOB_ID}"
    ;;
    criv1alpha2test)
      echo "run criv1alpha2 test"
      run_criv1alpha2_test
    ;;
    nodee2etest)
      echo "run node e2e test"
      run_node_e2e_test
    ;;
    *)
      echo "Unsupport action"
      exit 1
    ;;
  esac
}

main

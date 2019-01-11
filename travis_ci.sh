#!/bin/bash

set -e # fast fail

# action for travis ci
Action=$1

# config the number of integration test job.
JOBS=$2

# config the id of job
JOB_ID=$3

run_pre_test() {
    local caselist="test/testcase.list"

    # get the number of all test case
    # test case must write with this format:
    # func (XXX) TestXXX(c *check.C) {
    grep -h -E "^func\ \(.*\)\ Test.*\(c\ \*check\.C\)\ \{" ./test -r \
        | awk '{print $3,$4}' | awk -F \( '{print $1}' \
        | sed 's/\*//' | sed 's/) /./' > ${caselist}

    # make test case list file
    local sum nums
    local index=1
    local loop=1

    # shellcheck disable=SC2002
    sum=$(cat "${caselist}" | wc -l)
    nums=$((sum / JOBS + 1))

    rm -rf ${caselist}.*
    # shellcheck disable=SC2013
    for test in $(cat ${caselist}); do
        tmp=$((loop * nums))
        if [[ ${index} -gt ${tmp} ]]; then
            loop=$((loop + 1))
        fi
        index=$((index + 1))
        echo "${test}" >> "${caselist}.${loop}"
    done
}

run_unittest() {
    sudo env "PATH=$PATH" hack/install/install_ci_related.sh
    make unit-test
    make coverage
}

run_integration_test() {
    local job_id=$1

    make build
    make build-daemon-integration
    sudo env "PATH=$PATH" make install

    sudo env "PATH=$PATH" make download-dependencies
    sudo env "PATH=$PATH" "INTEGRATION_FLAGS=${job_id}" make integration-test
    make coverage
}

run_criv1alpha1_test() {
    make build
    TEST_FLAGS="" BUILDTAGS="selinux seccomp apparmor" make build-daemon-integration
    sudo env "PATH=$PATH" make install

    sudo env "PATH=$PATH" make download-dependencies
    sudo env "PATH=$PATH" make cri-v1alpha1-test
    make coverage
}

run_criv1alpha2_test() {
    make build
    TEST_FLAGS="" BUILDTAGS="selinux seccomp apparmor" make build-daemon-integration
    sudo env "PATH=$PATH" make install

    sudo env "PATH=$PATH" make download-dependencies
    sudo env "PATH=$PATH" make cri-v1alpha2-test
    make coverage
}

run_node_e2e_test() {
    make build
    TEST_FLAGS="" make build-daemon-integration
    sudo env "PATH=$PATH" make install

    sudo env "PATH=$PATH" make download-dependencies
    sudo env "PATH=$PATH" make cri-e2e-test
    make coverage
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
        criv1alpha1test)
            echo "run criv1alpha1 test"
            run_criv1alpha1_test
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
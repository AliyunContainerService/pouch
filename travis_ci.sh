#!/bin/bash

set -e # fast fail

Action=$1

run_unittest() {
    sudo env "PATH=$PATH" hack/install/install_ci_related.sh
    make unit-test
    make coverage
}

run_integration_test() {
    make build
    make build-daemon-integration
    TEST_FLAGS="" make build-integration-test
    sudo env "PATH=$PATH" make install

    sudo env "PATH=$PATH" make download-dependencies
    sudo env "PATH=$PATH" make integration-test
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
        unittest)
            echo "run unit test"
            run_unittest
        ;;
        integrationtest)
            echo "run integration test"
#            run_integration_test
        ;;
        criv1alpha1test)
            echo "run criv1alpha1 test"
#            run_criv1alpha1_test
        ;;
        criv1alpha2test)
            echo "run criv1alpha2 test"
#            run_criv1alpha2_test
        ;;
        nodee2etest)
            echo "run node e2e test"
#            run_node_e2e_test
        ;;
        *)
            echo "Unsupport action"
            exit 1
        ;;
    esac
}

main
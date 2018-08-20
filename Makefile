# TEST_FLAGS used as flags of go test.
TEST_FLAGS ?= -v --race

# DAEMON_BINARY_NAME is the name of binary of daemon.
DAEMON_BINARY_NAME=pouchd

# CLI_BINARY_NAME is the name of binary of pouch client.
CLI_BINARY_NAME=pouch

# DAEMON_INTEGRATION_BINARY_NAME is the name of test binary of daemon.
DAEMON_INTEGRATION_BINARY_NAME=pouchd-integration

# INTEGRATION_TESTCASE_BINARY_NAME is the name of binary of integration cases.
INTEGRATION_TESTCASE_BINARY_NAME=pouchd-integration-test

# DEST_DIR is base path used to install pouch & pouchd
DEST_DIR=/usr/local

# the following variables used for the daemon build

# API_VERSION is used for daemon API Version in go build.
API_VERSION="1.24"

# VERSION is used for daemon Release Version in go build.
VERSION ?= "1.0.0-rc1"

# GIT_COMMIT is used for daemon GitCommit in go build.
GIT_COMMIT=$(shell git describe --dirty --always --tags 2> /dev/null || true)

# BUILD_TIME is used for daemon BuildTime in go build.
BUILD_TIME=$(shell date --rfc-3339 s 2> /dev/null | sed -e 's/ /T/')

VERSION_PKG=github.com/alibaba/pouch
DEFAULT_LDFLAGS="-X ${VERSION_PKG}/version.GitCommit=${GIT_COMMIT} \
		  -X ${VERSION_PKG}/version.Version=${VERSION} \
		  -X ${VERSION_PKG}/version.ApiVersion=${API_VERSION} \
		  -X ${VERSION_PKG}/version.BuildTime=${BUILD_TIME}"

# COVERAGE_PACKAGES is the coverage we care about.
COVERAGE_PACKAGES=$(shell go list ./... | \
				  grep -v github.com/alibaba/pouch$$ | \
				  grep -v github.com/alibaba/pouch/storage/volume/examples/demo | \
				  grep -v github.com/alibaba/pouch/test | \
				  grep -v github.com/alibaba/pouch/cli | \
				  grep -v github.com/alibaba/pouch/cri/apis | \
				  grep -v github.com/alibaba/pouch/apis/types )

COVERAGE_PACKAGES_LIST=$(shell echo $(COVERAGE_PACKAGES) | tr " " ",")

build: build-daemon build-cli ## build PouchContainer both daemon and cli binaries

build-daemon: modules ## build PouchContainer daemon binary
	@echo "$@: bin/${DAEMON_BINARY_NAME}"
	@mkdir -p bin
	@GOOS=linux go build -ldflags ${DEFAULT_LDFLAGS} -o bin/${DAEMON_BINARY_NAME} -tags 'selinux'

build-cli: ## build PouchContainer cli binary
	@echo "$@: bin/${CLI_BINARY_NAME}"
	@mkdir -p bin
	@go build -o bin/${CLI_BINARY_NAME} github.com/alibaba/pouch/cli

build-daemon-integration: modules ## build PouchContainer daemon integration testing binary
	@echo $@
	@mkdir -p bin
	go test -c ${TEST_FLAGS} \
		-cover -covermode=atomic -coverpkg ${COVERAGE_PACKAGES_LIST} \
		-o bin/${DAEMON_INTEGRATION_BINARY_NAME}

build-integration-test: modules ## build PouchContainer integration test-case binary
	@echo $@
	@mkdir -p bin
	go test -c \
		-o bin/${INTEGRATION_TESTCASE_BINARY_NAME} github.com/alibaba/pouch/test

modules: ## run modules to generate volume related code
	@echo "build volume $@"
	@./hack/module --clean
	@./hack/module --add-volume=github.com/alibaba/pouch/storage/volume/modules/tmpfs
	@./hack/module --add-volume=github.com/alibaba/pouch/storage/volume/modules/local

install: ## install pouch and pouchd binary into /usr/local/bin
	@echo $@
	@mkdir -p $(DEST_DIR)/bin
	install bin/$(CLI_BINARY_NAME) $(DEST_DIR)/bin
	install bin/$(DAEMON_BINARY_NAME) $(DEST_DIR)/bin

uninstall: ## uninstall pouchd and pouch binary
	@echo $@
	@rm -f $(addprefix $(DEST_DIR)/bin/,$(notdir $(DAEMON_BINARY_NAME)))
	@rm -f $(addprefix $(DEST_DIR)/bin/,$(notdir $(CLI_BINARY_NAME)))

.PHONY: package-dependencies
package-dependencies: ## install containerd, runc and lxcfs dependencies for packaging
	@echo $@
	hack/install/install_containerd.sh
	hack/install/install_lxcfs.sh
	hack/install/install_runc.sh

.PHONY: download-dependencies
download-dependencies: package-dependencies ## install dumb-init, local-persist, nsenter and CI tools dependencies
	@echo $@
	hack/install/install_ci_related.sh
	hack/install/install_dumb_init.sh
	hack/install/install_local_persist.sh
	hack/install/install_nsenter.sh

.PHONY: clean
clean: ## clean to remove bin/* and files created by module
	@go clean
	@rm -f bin/*
	@rm -rf coverage/*
	@./hack/module --clean


.PHONY: check
check: gometalinter validate-swagger ## run all linters

.PHONY: validate-swagger
validate-swagger: ## run swagger validate
	@echo $@
	./hack/validate_swagger.sh

# gometalinter consumes .gometalinter.json as config.
.PHONY: gometalinter
gometalinter: ## run gometalinter for go source code
	@echo $@
	gometalinter --config .gometalinter.json ./...


.PHONY: unit-test
unit-test: modules ## run go unit-test
	@echo $@
	@mkdir -p coverage
	@( for pkg in ${COVERAGE_PACKAGES}; do \
		go test ${TEST_FLAGS} \
			-cover -covermode=atomic \
			-coverprofile=coverage/unit-test-`echo $$pkg | tr "/" "_"`.out \
			$$pkg || exit; \
	done )

.PHONY: integration-test
integration-test: ## run daemon integration-test
	@echo $@
	@mkdir -p coverage
	./hack/testing/run_daemon_integration.sh

.PHONY: cri-v1alpha1-test
cri-v1alpha1-test: ## run v1 alpha1 cri-v1alpha1-test
	@echo $@
	@mkdir -p coverage
	./hack/testing/run_daemon_cri_integration.sh v1alpha1

.PHONY: cri-v1alpha2-test
cri-v1alpha2-test: ## run v1 alpha2 cri-v1alpha2-test
	@echo $@
	@mkdir -p coverage
	./hack/testing/run_daemon_cri_integration.sh v1alpha2

.PHONY: test
test: unit-test integration-test cri-v1alpha1-test cri-v1alpha2-test ## run the unit-test, integration-test , cri-v1alpha1-test and cri-v1alpha2-test

.PHONY: coverage
coverage: ## combine coverage after test
	@echo $@
	@gocovmerge coverage/* > coverage.txt


.PHONY: help
help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-28s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

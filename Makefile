# Copyright The PouchContainer Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# TEST_FLAGS used as flags of go test.
TEST_FLAGS ?= -v --race

# INTEGRATION_FLAGS used as flags of run integration test
INTEGRATION_FLAGS ?= ""

# DAEMON_BINARY_NAME is the name of binary of daemon.
DAEMON_BINARY_NAME=pouchd

# CLI_BINARY_NAME is the name of binary of pouch client.
CLI_BINARY_NAME=pouch

# DAEMON_INTEGRATION_BINARY_NAME is the name of test binary of daemon.
DAEMON_INTEGRATION_BINARY_NAME=pouchd-integration

# INTEGRATION_TESTCASE_BINARY_NAME is the name of binary of integration cases.
INTEGRATION_TESTCASE_BINARY_NAME=pouchd-integration-test

# GOARCH is the target platform for build
GOARCH ?= $(shell go env GOARCH)
GOPATH ?= $(shell go env GOPATH)

# CC is the cross compiler
ifeq (${GOARCH},arm64)
	CC=aarch64-linux-gnu-gcc
else ifeq (${GOARCH},ppc64le)
	CC=powerpc64le-linux-gnu-gcc
else
	GOARCH=amd64
endif

# BUILD_ROOT is specified
BUILD_ROOT := ${CURDIR}/install/${GOARCH}

# DEST_DIR is base path used to install pouch & pouchd
DEST_DIR=/usr/local

# PREFIX is base path to install pouch & pouchd
# PREFIX will override the value of DEST_DIR when specified
# example: make install PREFIX=/usr
ifdef PREFIX
	DEST_DIR := $(PREFIX)
endif

# the following variables used for the daemon build

# API_VERSION is used for daemon API Version in go build.
API_VERSION="1.24"

# VERSION is used for daemon Release Version in go build.
VERSION ?= "1.2.0"

# GIT_COMMIT is used for daemon GitCommit in go build.
GIT_COMMIT=$(shell git describe --dirty --always --tags 2> /dev/null || true)

# BUILD_TIME is used for daemon BuildTime in go build.
BUILD_TIME=$(shell date --rfc-3339 s 2> /dev/null | sed -e 's/ /T/')

VERSION_PKG=github.com/alibaba/pouch
DEFAULT_LDFLAGS="-X ${VERSION_PKG}/version.GitCommit=${GIT_COMMIT} \
		  -X ${VERSION_PKG}/version.Version=${VERSION} \
		  -X ${VERSION_PKG}/version.ApiVersion=${API_VERSION} \
		  -X ${VERSION_PKG}/version.BuildTime=${BUILD_TIME}"

GOBUILD_TAGS=$(if $(BUILDTAGS),-tags "$(BUILDTAGS)",)

# COVERAGE_PACKAGES is the coverage we care about.
COVERAGE_PACKAGES=$(shell go list ./... | \
				  grep -v github.com/alibaba/pouch$$ | \
				  grep -v github.com/alibaba/pouch/storage/volume/examples/demo | \
				  grep -v github.com/alibaba/pouch/test | \
				  grep -v github.com/alibaba/pouch/cli$$ | \
				  grep -v github.com/alibaba/pouch/cli/ | \
				  grep -v github.com/alibaba/pouch/cri/apis | \
				  grep -v github.com/alibaba/pouch/cri/v1alpha1 | \
				  grep -v github.com/alibaba/pouch/apis/types )

COVERAGE_PACKAGES_LIST=$(shell echo $(COVERAGE_PACKAGES) | tr " " ",")

#  POUCH IMAGE is the develop image for cross-building
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
GIT_BRANCH_CLEAN := $(shell echo $(GIT_BRANCH) | sed -e "s/[^[:alnum:]]/-/g")
POUCH_IMAGE := pouch_dev$(if $(GIT_BRANCH_CLEAN),:$(GIT_BRANCH_CLEAN))

# Project location
RUNC_PRO := github.com/opencontainers/runc
CONTAINERD_PRO := github.com/containerd/containerd
POUCH_PRO := github.com/alibaba/pouch
LXCFS_PRO := github.com/lxc/lxcfs

# Runc cross building configuration
RUNC_VERSION ?= "v1.0.0-rc6-1"
RUNC_BUILDTAGS := seccomp
RUNC_EXTRA_FLAGS :=
RUNC_COMMIT_NO :=
RUNC_COMMIT :=

# CONTAINERD cross-building configuration
CONTAINERD_VERSION := "v1.0.3"

# LXCFS cross building configuration
LXCFS_VERSION := "stable-2.0"

build: build-daemon build-cli ## build PouchContainer both daemon and cli binaries

build-daemon: modules plugin ## build PouchContainer daemon binary
	@echo "$@: bin/${DAEMON_BINARY_NAME}"
	@mkdir -p bin
	@GOOS=linux go build -ldflags ${DEFAULT_LDFLAGS} ${GOBUILD_TAGS} -o bin/${DAEMON_BINARY_NAME}

build-cli: ## build PouchContainer cli binary
	@echo "$@: bin/${CLI_BINARY_NAME}"
	@mkdir -p bin
	@go build -o bin/${CLI_BINARY_NAME} github.com/alibaba/pouch/cli

dev-image: ## build the Docker Image as cross building environment
	docker build -f Dockerfile.${GOARCH}.cross . -t ${POUCH_IMAGE}

shell: dev-image ## enter into the cross building shell environment
	docker run -ti --privileged --rm -v ${CURDIR}:/go/src/${POUCH_PRO} \
		${POUCH_IMAGE} \
		bash

download-source: ## download source code for cross-building
	@echo "remove the source code"
	rm -rf ${GOPATH}/src/${RUNC_PRO}
	rm -rf ${GOPATH}/src/${CONTAINERD_PRO}
	rm -rf ${GOPATH}/src/${LXCFS_PRO}
	mkdir -p ${GOPATH}/src/${RUNC_PRO}
	mkdir -p ${GOPATH}/src/${CONTAINERD_PRO}
	mkdir -p ${GOPATH}/src/${LXCFS_PRO}
	@echo "download the runc source code"
	git clone -b ${RUNC_VERSION} \
		https://github.com/alibaba/runc \
		${GOPATH}/src/${RUNC_PRO}
	@echo "download the containerd source code"
	git clone -b ${CONTAINERD_VERSION} \
		https://${CONTAINERD_PRO} \
		${GOPATH}/src/${CONTAINERD_PRO}
	@echo "download the containerd source code"
	git clone -b ${LXCFS_VERSION} \
		https://${LXCFS_PRO} \
		${GOPATH}/src/${LXCFS_PRO}

cross-build: dev-image ## cross build pouchd pouchd runc containerd on host
	docker run -e BUILDTAGS="$(BUILDTAGS)" -e GOARCH="${GOARCH}" \
		--rm -v ${CURDIR}:/go/src/${POUCH_PRO} \
		${POUCH_IMAGE} \
		make local-cross

local-cross: modules plugin download-source ## local cross build pouch pouchd runc containerd inside container
	@echo "$@: ${BUILD_ROOT}/bin/${DAEMON_BINARY_NAME}"
	@mkdir -p ${BUILD_ROOT}/bin
	@CGO_ENABLED=1 GOARCH=${GOARCH} GOOS=linux CC=${CC} \
		go build -ldflags ${DEFAULT_LDFLAGS} ${GOBUILD_TAGS} \
		-o ${BUILD_ROOT}/bin/${DAEMON_BINARY_NAME}
	@echo "$@: ${BUILD_ROOT}/bin/${CLI_BINARY_NAME}"
	@CGO_ENABLED=0 GOARCH=${GOARCH} \
		go build -o ${BUILD_ROOT}/bin/${CLI_BINARY_NAME} \
		github.com/alibaba/pouch/cli
	@echo "$@: ${BUILD_ROOT}/bin/runc"
	@CGO_ENABLED=1 GOARCH=${GOARCH} GOOS=linux CC=${CC} \
		go build -buildmode=pie ${RUNC_EXTRA_FLAGS} \
		-ldflags "-X main.gitCommit=${RUNC_COMMIT} -X main.version=${RUNC_VERSION} ${RUNC_EXTRA_LDFLAGS}" \
		-tags "${RUNC_BUILDTAGS}" \
		-o ${BUILD_ROOT}/bin/runc ${RUNC_PRO}
	@CGO_ENABLED=1 GOARCH=${GOARCH} GOOS=linux CC=${CC} \
		${MAKE} -C /go/src/${CONTAINERD_PRO} && cp -Rf /go/src/${CONTAINERD_PRO}/bin/* ${BUILD_ROOT}/bin
	@echo "cross building lxcfs"
	cd /go/src/${LXCFS_PRO} \
		&& grep -l -r "liblxcfs" . | xargs sed -i 's/liblxcfs/libpouchlxcfs/g' \
		&& ./bootstrap.sh
ifeq (${GOARCH},amd64)
	cd /go/src/${LXCFS_PRO} \
		&& ./configure --prefix=${BUILD_ROOT}
endif
ifeq (${GOARCH},arm64)
	cd /go/src/${LXCFS_PRO} \
		&& ./configure --prefix=${BUILD_ROOT} \
		--host=aarch64-linux-gnu \
		--with-gnu-ld \
		--with-distro=redhat
endif
ifeq (${GOARCH},ppc64le)
	cd /go/src/${LXCFS_PRO} \
		&& ./configure --prefix=${BUILD_ROOT} \
		--host=powerpc64le-linux-gnu \
		--with-gnu-ld \
		--with-distro=redhat
endif
	@cd /go/src/${LXCFS_PRO} && ${MAKE} clean && ${MAKE} && ${MAKE} install
	@mv ${BUILD_ROOT}/bin/lxcfs ${BUILD_ROOT}/bin/pouch-lxcfs

build-daemon-integration: modules plugin ## build PouchContainer daemon integration testing binary
	@echo $@
	@mkdir -p bin
	go test -c ${TEST_FLAGS} ${GOBUILD_TAGS} \
		-cover -covermode=atomic -coverpkg ${COVERAGE_PACKAGES_LIST} \
		-o bin/${DAEMON_INTEGRATION_BINARY_NAME}

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
	hack/install/install_criu.sh

.PHONY: clean
clean: ## clean to remove bin/* and files created by module
	@go clean
	@rm -f bin/*
	@rm -rf install/*
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
unit-test: modules plugin ## run go unit-test
	@echo $@
	@mkdir -p coverage
	@( for pkg in ${COVERAGE_PACKAGES}; do \
		go test ${TEST_FLAGS} \
			-cover -covermode=atomic \
			-coverprofile=coverage/unit-test-`echo $$pkg | tr "/" "_"`.out \
			$$pkg || exit; \
	done )

.PHONY: integration-test
integration-test: build-daemon-integration ## run daemon integration-test
	@echo $@
	@mkdir -p coverage
	./hack/testing/run_daemon_integration.sh ${INTEGRATION_FLAGS}

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

.PHONY: cri-e2e-test
cri-e2e-test: ## run cri-e2e-test
	@echo $@
	@mkdir -p coverage
	./hack/testing/run_daemon_cri_e2e.sh v1alpha2

.PHONY: test
test: unit-test integration-test cri-v1alpha1-test cri-v1alpha2-test cri-e2e-test ## run the unit-test, integration-test , cri-v1alpha1-test , cri-v1alpha2-test and cri-e2e-test

.PHONY: coverage
coverage: ## combine coverage after test
	@echo $@
	@gocovmerge coverage/* > coverage.txt

.PHONY: plugin
plugin: ## build hook plugin
	@echo "build $@"
	@./hack/module --add-plugin=github.com/alibaba/pouch/hookplugins/containerplugin
	@./hack/module --add-plugin=github.com/alibaba/pouch/hookplugins/daemonplugin
	@./hack/module --add-plugin=github.com/alibaba/pouch/hookplugins/criplugin
	@./hack/module --add-plugin=github.com/alibaba/pouch/hookplugins/volumeplugin
	@./hack/module --add-plugin=github.com/alibaba/pouch/hookplugins/apiplugin
	@./hack/module --add-plugin=github.com/alibaba/pouch/hookplugins/imageplugin

.PHONY: help
help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-28s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

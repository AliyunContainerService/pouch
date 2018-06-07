# Go parameters
GOBUILD=go build
GOCLEAN=go clean
GOTEST=go test
GOPACKAGES=$(shell go list ./... | grep -v /vendor/ | sed 's/^_//')

# Binary name of CLI and Daemon
BINARY_NAME=pouchd
CLI_BINARY_NAME=pouch

# Base path used to install pouch & pouchd
DESTDIR=/usr/local

.PHONY: build
build: server client

.PHONY: pre
pre:
	@./hack/build pre

.PHONY: server
server: pre modules
	@./hack/build server

.PHONY: client
client: pre
	@./hack/build client

.PHONY: testserver
testserver: pre modules
	@./hack/build testserver

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(CLI_BINARY_NAME)
	./hack/build clean
	./hack/module --clean

.PHONY: check
check: pre fmt lint vet validate-swagger

.PHONY: fmt
fmt: ## run go fmt
	@echo $@
	@which gofmt
	@test -z "$$(gofmt -s -l . 2>/dev/null | grep -Fv 'vendor/' | grep -v ".pb.go$$" | tee /dev/stderr)" || \
		(echo "please format Go code with 'gofmt -s -w'" && false)
	@test -z "$$(find . -path ./vendor -prune -o ! -name timestamp.proto ! -name duration.proto -name '*.proto' -type f -exec grep -Hn -e "^ " {} \; | tee /dev/stderr)" || \
		(echo "please indent proto files with tabs only" && false)
	@test -z "$$(find . -path ./vendor -prune -o -name '*.proto' -type f -exec grep -Hn "Meta meta = " {} \; | grep -v '(gogoproto.nullable) = false' | tee /dev/stderr)" || \
		(echo "meta fields in proto files must have option (gogoproto.nullable) = false" && false)

.PHONY: lint
lint: ## run go lint
	@echo $@
	@which golint
	@test -z "$$(golint ./... | grep -Fv 'vendor/' | grep -v ".pb.go:" | tee /dev/stderr)"

.PHONY: vet
vet: ## run go vet
	@echo $@
	@test -z "$$(./hack/build vet)"

.PHONY: unit-test
unit-test: pre modules ## run go test
	@echo $@
	@./hack/build unit-test

.PHONY: validate-swagger
validate-swagger: ## run swagger validate
	@echo $@
	@swagger validate apis/swagger.yml

.PHONY: modules
modules:
	@./hack/module --clean
	@./hack/module --add-volume=github.com/alibaba/pouch/storage/volume/modules/tmpfs
	@./hack/module --add-volume=github.com/alibaba/pouch/storage/volume/modules/local

# build binaries
# install them to /usr/local/bin/
# remove binaries
.PHONY: install
install: build ## build and install binary into /usr/local/bin
	@echo $@
	@echo "installing $(BINARY_NAME) and $(CLI_BINARY_NAME) to $(DESTDIR)/bin"
	@mkdir -p $(DESTDIR)/bin
	@install $(BINARY_NAME) $(DESTDIR)/bin
	@install $(CLI_BINARY_NAME) $(DESTDIR)/bin

.PHONY: uninstall
uninstall: ## uninstall pouchd and pouch binary
	@echo $@
	@rm -f $(addprefix $(DESTDIR)/bin/,$(notdir $(BINARY_NAME)))
	@rm -f $(addprefix $(DESTDIR)/bin/,$(notdir $(CLI_BINARY_NAME)))

# For integration-test and test, PATH is not set under sudo, then we set up path mannually.
# Ref https://unix.stackexchange.com/questions/83191/how-to-make-sudo-preserve-path
.PHONY: integration-test
integration-test: ## build binary and run integration-test
	@bash -c "env PATH=$(PATH) hack/make.sh build integration-test"

.PHONY: cri-test
cri-test: ## build binary and run cri-test
	@bash -c "env PATH=$(PATH) hack/make.sh build cri-test"

.PHONY: test
test: ## run the build integration-test cri-test
	@bash -c "env PATH=$(PATH) hack/make.sh build integration-test cri-test"

.PHONY: help
help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

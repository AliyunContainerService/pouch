# Go parameters
GOBUILD=go build
GOCLEAN=go clean
GOTEST=go test
GOPATH=$(shell go env GOPATH)
GOPACKAGES=$(shell go list ./... | grep -v /vendor/ | sed 's/^_//')

# Binary name of CLI and Daemon
BINARY_NAME=pouchd
CLI_BINARY_NAME=pouch

# Base path used to install pouch & pouchd
DESTDIR=/usr/local

.PHONY: build
build: server client

.PHONY: server
server: modules
	GOOS=linux $(GOBUILD) -o $(BINARY_NAME)

.PHONY: client
client:
	$(GOBUILD) -o $(CLI_BINARY_NAME) github.com/alibaba/pouch/cli

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(CLI_BINARY_NAME)
	./module --clean

.PHONY: check
check: fmt lint vet validate-swagger

.PHONY: fmt
fmt: ## run go fmt
	@echo $@
	@test -z "$$(gofmt -s -l . 2>/dev/null | grep -Fv 'vendor/' | grep -v ".pb.go$$" | tee /dev/stderr)" || \
		(echo "please format Go code with 'gofmt -s -w'" && false)
	@test -z "$$(find . -path ./vendor -prune -o ! -name timestamp.proto ! -name duration.proto -name '*.proto' -type f -exec grep -Hn -e "^ " {} \; | tee /dev/stderr)" || \
		(echo "please indent proto files with tabs only" && false)
	@test -z "$$(find . -path ./vendor -prune -o -name '*.proto' -type f -exec grep -Hn "Meta meta = " {} \; | grep -v '(gogoproto.nullable) = false' | tee /dev/stderr)" || \
		(echo "meta fields in proto files must have option (gogoproto.nullable) = false" && false)

.PHONY: lint
lint: ## run go lint
	@echo $@
	@test -z "$$(golint ./... | grep -Fv 'vendor/' | grep -v ".pb.go:" | tee /dev/stderr)"

.PHONY: vet
vet: # run go vet
	@echo $@
	@test -z "$$(go vet ${GOPACKAGES} 2>&1 | grep -v "unrecognized printf verb 'r'" | egrep -v '(exit status 1)' | tee /dev/stderr)"

.PHONY: unit-test
unit-test: ## run go test
	@echo $@
	@go test `go list ./... | grep -v 'github.com/alibaba/pouch/test'`

.PHONY: validate-swagger
validate-swagger: ## run swagger validate
	@echo $@
	@swagger validate apis/swagger.yml

.PHONY: modules
modules:
	@./module --add-volume=github.com/alibaba/pouch/volume/modules/ceph
	@./module --add-volume=github.com/alibaba/pouch/volume/modules/tmpfs
	@./module --add-volume=github.com/alibaba/pouch/volume/modules/local
	
# build binaries
# install them to /usr/local/bin/
# remove binaries
.PHONY: install
install: build
	@echo $@
	@echo "installing $(BINARY_NAME) and $(CLI_BINARY_NAME) to $(DESTDIR)/bin"
	@mkdir -p $(DESTDIR)/bin
	@install $(BINARY_NAME) $(DESTDIR)/bin
	@install $(CLI_BINARY_NAME) $(DESTDIR)/bin

.PHONY: uninstall
uninstall:
	@echo $@
	@rm -f $(addprefix $(DESTDIR)/bin/,$(notdir $(BINARY_NAME)))
	@rm -f $(addprefix $(DESTDIR)/bin/,$(notdir $(CLI_BINARY_NAME)))

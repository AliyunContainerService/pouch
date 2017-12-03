# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOPATH=$(shell cd ../../../..; pwd)
GOPACKAGES=$(shell GOPATH=`cd ../../../..; pwd` go list ./... | grep -v /vendor/ | sed 's/^_//')

# Binary name of CLI and Daemon
BINARY_NAME=pouchd
CLI_BINARY_NAME=pouch

# Base path used to install pouch & pouchd
DESTDIR=/usr/local

build: server client

server: modules
	GOOS=linux $(GOBUILD) -o $(BINARY_NAME)

client:
	$(GOBUILD) -o $(CLI_BINARY_NAME) github.com/alibaba/pouch/cli

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(CLI_BINARY_NAME)
	./module --clean

check: fmt lint vet validate-swagger

fmt: ## run go fmt
	@echo $@
	@test -z "$$(gofmt -s -l . 2>/dev/null | grep -Fv 'vendor/' | grep -v ".pb.go$$" | tee /dev/stderr)" || \
		(echo "please format Go code with 'gofmt -s -w'" && false)
	@test -z "$$(find . -path ./vendor -prune -o ! -name timestamp.proto ! -name duration.proto -name '*.proto' -type f -exec grep -Hn -e "^ " {} \; | tee /dev/stderr)" || \
		(echo "please indent proto files with tabs only" && false)
	@test -z "$$(find . -path ./vendor -prune -o -name '*.proto' -type f -exec grep -Hn "Meta meta = " {} \; | grep -v '(gogoproto.nullable) = false' | tee /dev/stderr)" || \
		(echo "meta fields in proto files must have option (gogoproto.nullable) = false" && false)

lint: ## run go lint
	@echo $@
	@test -z "$$(golint ./... | grep -Fv 'vendor/' | grep -v ".pb.go:" | tee /dev/stderr)"

vet: # run go vet
	@echo $@
	@test -z "$$(go vet ${GOPACKAGES} 2>&1 | grep -v "unrecognized printf verb 'r'" | egrep -v '(exit status 1)' | tee /dev/stderr)"

unit-test: ## run go test
	@echo $@
	@go test `go list ./... | grep -v 'github.com/alibaba/pouch/test'`

validate-swagger: ## run swagger validate
	@echo $@
	@swagger validate apis/swagger.yml

modules:
	@./module --add-volume=github.com/alibaba/pouch/volume/modules/ceph
	@./module --add-volume=github.com/alibaba/pouch/volume/modules/tmpfs
	@./module --add-volume=github.com/alibaba/pouch/volume/modules/local
	
# build binaries
# install them to /usr/local/bin/
# remove binaries
install: build
	@echo $@
	@echo "installing $(BINARY_NAME) and $(CLI_BINARY_NAME) to $(DESTDIR)/bin"
	@mkdir -p $(DESTDIR)/bin
	@install $(BINARY_NAME) $(DESTDIR)/bin
	@install $(CLI_BINARY_NAME) $(DESTDIR)/bin

uninstall:
	@echo $@
	@rm -f $(addprefix $(DESTDIR)/bin/,$(notdir $(BINARY_NAME)))
	@rm -f $(addprefix $(DESTDIR)/bin/,$(notdir $(CLI_BINARY_NAME)))
	
.PHONY: check \
  build \
  client \
  server \
  clean \
  fmt \
  lint \
  vet \
  unit-test \
  modules \
  validate-swagger \
  install \
  uninstall

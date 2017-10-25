# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=pouchd
CLI_BINARY_NAME=pouch
GOPATH=$(shell cd ../../../..; pwd)
PACKAGES=$(shell GOPATH=`cd ../../../..; pwd` go list ./... | grep -v /vendor/ | sed 's/^_//')

build: server client

server: modules
	GOOS=linux $(GOBUILD) -o $(BINARY_NAME)

client:
	GOOS=linux $(GOBUILD) -o $(CLI_BINARY_NAME) github.com/alibaba/pouch/cli

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(CLI_BINARY_NAME)
	./module --clean

check: fmt lint vet

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
	@test -z "$$(go vet ${PACKAGES} 2>&1 | tee /dev/stderr)"

unit-test: ## run go test
	@echo $@
	@go test $(go list ./... | grep -v /vendor/)

modules:
	@./module --add-volume=github.com/alibaba/pouch/volume/modules/ceph
	@./module --add-volume=github.com/alibaba/pouch/volume/modules/tmpfs
	@./module --add-volume=github.com/alibaba/pouch/volume/modules/local

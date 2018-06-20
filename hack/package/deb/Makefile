# Go parameters
GOBUILD=go build
GOCLEAN=go clean
GOTEST=go test
GOPACKAGES=$(shell go list ./... | grep -v /vendor/ | sed 's/^_//')

# Binary name of CLI and Daemon
BINARY_NAME=pouchd
CLI_BINARY_NAME=pouch

# Base path used to install pouch & pouchd
DESTDIR=debian/pouch/usr/local/bin/
TMP=/tmp/tmp.pouch

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

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(CLI_BINARY_NAME)
	rm -rf debian/pouch/*
	./hack/build clean
	./hack/module --clean

#.PHONY: modules
modules:
	@./hack/module --add-volume=github.com/alibaba/pouch/storage/volume/modules/ceph
	@./hack/module --add-volume=github.com/alibaba/pouch/storage/volume/modules/tmpfs
	@./hack/module --add-volume=github.com/alibaba/pouch/storage/volume/modules/local

# install binaries
.PHONY: install
install: build
	@echo $@
	mkdir -p debian/pouch/usr/bin/
	cp pouch debian/pouch/usr/bin/
	cp pouchd debian/pouch/usr/bin/
	# install containerd
	echo "Download and install containerd."
	[ -d $(TMP) ] || mkdir -p /tmp/tmp.pouch
	wget --quiet https://github.com/containerd/containerd/releases/download/v1.0.3/containerd-1.0.3.linux-amd64.tar.gz -P $(TMP)
	tar xf $(TMP)/containerd-1.0.3.linux-amd64.tar.gz -C $(TMP) && cp -f $(TMP)/bin/* debian/pouch/usr/bin/

	# install runc
	echo "Download and install runc."
	wget --quiet https://github.com/alibaba/runc/releases/download/v1.0.0-rc4-1/runc.amd64 -P $(TMP)/bin
	chmod +x $(TMP)/bin/runc.amd64
	mv $(TMP)/bin/runc.amd64 debian/pouch/usr/bin/runc

	# pouch support 'systemctl' command
	mkdir -p debian/pouch/usr/lib/systemd/system/
	cp hack/package/deb/systemd/pouch.service debian/pouch/usr/lib/systemd/system/

.PHONY: uninstall
uninstall:
	@echo $@
	@rm -f $(addprefix $(DESTDIR)/bin/,$(notdir $(BINARY_NAME)))
	@rm -f $(addprefix $(DESTDIR)/bin/,$(notdir $(CLI_BINARY_NAME)))

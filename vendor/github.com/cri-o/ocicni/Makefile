GO ?= go
EPOCH_TEST_COMMIT ?= b0fc980
PROJECT := github.com/cri-o/ocicni
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
GIT_BRANCH_CLEAN := $(shell echo $(GIT_BRANCH) | sed -e "s/[^[:alnum:]]/-/g")
OCICNI_IMAGE := ocicni_dev$(if $(GIT_BRANCH_CLEAN),:$(GIT_BRANCH_CLEAN))
OCICNI_INSTANCE := ocicni_dev
PREFIX ?= ${DESTDIR}/usr/local
BINDIR ?= ${PREFIX}/bin
LIBEXECDIR ?= ${PREFIX}/libexec
MANDIR ?= ${PREFIX}/share/man
ETCDIR ?= ${DESTDIR}/etc

GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_INFO := $(shell date +%s)

# If GOPATH not specified, use one in the local directory
ifeq ($(GOPATH),)
export GOPATH := $(CURDIR)/_output
unexport GOBIN
endif
GOPKGDIR := $(GOPATH)/src/$(PROJECT)
GOPKGBASEDIR := $(shell dirname "$(GOPKGDIR)")

# Update VPATH so make finds .gopathok
VPATH := $(VPATH):$(GOPATH)

LDFLAGS := -ldflags '-X main.gitCommit=${GIT_COMMIT} -X main.buildInfo=${BUILD_INFO}'

all: binaries

.gopathok:
ifeq ("$(wildcard $(GOPKGDIR))","")
	mkdir -p "$(GOPKGBASEDIR)"
	ln -s "$(CURDIR)" "$(GOPKGBASEDIR)"
endif
	touch "$(GOPATH)/.gopathok"

gofmt:
	@./hack/verify-gofmt.sh

binaries: ocicnitool

ocicnitool: .gopathok $(shell hack/find-godeps.sh $(GOPKGDIR) tools/ocicnitool $(PROJECT))
	$(GO) build $(LDFLAGS) -tags "$(BUILDTAGS)" -o $@ $(PROJECT)/tools/ocicnitool

check: .gopathok
	@./hack/test-go.sh $(GOPKGDIR)
	@./hack/verify-gofmt.sh

clean:
ifneq ($(GOPATH),)
	rm -f "$(GOPATH)/.gopathok"
endif
	rm -rf _output

# When this is running in travis, it will only check the travis commit range
.gitvalidation: .gopathok
ifeq ($(TRAVIS),true)
	GIT_CHECK_EXCLUDE="./vendor ./_output" $(GOPATH)/bin/git-validation -q -run DCO,short-subject,dangling-whitespace
else
	GIT_CHECK_EXCLUDE="./vendor ./_output" $(GOPATH)/bin/git-validation -v -run DCO,short-subject,dangling-whitespace -range $(EPOCH_TEST_COMMIT)..HEAD
endif

install.tools: .install.gitvalidation

.install.gitvalidation: .gopathok
	if [ ! -x "$(GOPATH)/bin/git-validation" ]; then \
		go get -u github.com/vbatts/git-validation; \
	fi

.PHONY: \
	binaries \
	clean \
	default \
	gofmt \
	help \
	check \
	install.tools \
	.gitvalidation

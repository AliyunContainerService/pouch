FROM ubuntu:16.04

# install wget to download golang source code
# install git
RUN apt-get update \
    && apt-get install -y \
    wget \ 
    git \
    make \
    gcc \
    vim \
    tree \
    rubygems \
    && apt-get clean

# set go version this image use
ENV GO_VERSION=1.9.1
ENV ARCH=amd64

# install golang which version is GO_VERSION
RUN wget --quiet https://storage.googleapis.com/golang/go${GO_VERSION}.linux-${ARCH}.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-${ARCH}.tar.gz \
    && rm go${GO_VERSION}.linux-${ARCH}.tar.gz

# create GOPATH
RUN mkdir /go
ENV GOPATH=/go

# set go binary path to local $PATH
# go binary path is /usr/local/go/bin
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# install golint. currently golint has no released version.
RUN go get -u golang.org/x/lint/golint

# install swagger 0.12.0
RUN wget --quiet -O /bin/swagger https://github.com/go-swagger/go-swagger/releases/download/0.12.0/swagger_linux_amd64 \
    && chmod +x /bin/swagger

# install markdownlint tool
RUN gem install rake \
    && gem install bundler \
    && git clone https://github.com/markdownlint/markdownlint.git \
    && cd markdownlint \
    && git checkout v0.4.0 \
    && rake install

COPY . /go/src/github.com/alibaba/pouch

WORKDIR /go/src/github.com/alibaba/pouch

# The environment is setup, when run what you need, just setup the CMD when `pouch run`
# For example, this Dockerfile will build an image named `pouch-image`.
# When running unit test, just execute `pouch run pouch-image make unit-test`.

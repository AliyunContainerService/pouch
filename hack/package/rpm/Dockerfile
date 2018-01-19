FROM centos:7.2.1511

# install wget to download golang source code
# install git
RUN yum update -y \
    && yum install -y \
    wget \
    git \
    gcc \
    ruby-devel \
    fuse-devel \
    pam-devel \
    automake \
    autoconf \
    libtool \
    make \
    rpm-build \
    rpm-sign \
    rubygems \
    vim \
    tree \
    && yum clean all
# install fpm to build rpm package
RUN gem install --no-ri --no-rdoc fpm

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

RUN git clone https://github.com/zzchun/pouch.git /go/src/github.com/alibaba/pouch

# set go binary path to local $PATH
# go binary path is /usr/local/go/bin
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /go/src/github.com/alibaba/pouch
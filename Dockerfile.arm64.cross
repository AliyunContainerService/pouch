FROM golang:1.10.4-stretch

RUN dpkg --add-architecture arm64 \
    && apt-get update && apt-get install -y \
    vim \
    git \
    curl \
    automake \
    libtool \
    help2man \
    libseccomp2 \
    libseccomp-dev \
    crossbuild-essential-arm64 \
    libseccomp-dev:arm64\
    libfuse-dev:arm64 \
    libpam-dev:arm64 \
    lxcfs \
    btrfs-progs \
    --no-install-recommends \
    && apt-get clean

ADD . /go/src/github.com/alibaba/pouch
WORKDIR /go/src/github.com/alibaba/pouch

FROM golang:1.10.4-stretch

RUN apt-get update && apt-get install -y \
    vim \
    git \
    build-essential \
    curl \
    automake \
    libtool \
    help2man \
    libseccomp2 \
    libseccomp-dev \
    libfuse-dev \
    libpam-dev \
    lxcfs \
    btrfs-progs \
    --no-install-recommends \
    && apt-get clean

ADD . /go/src/github.com/alibaba/pouch
WORKDIR /go/src/github.com/alibaba/pouch

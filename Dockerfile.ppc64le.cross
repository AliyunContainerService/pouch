FROM golang:1.10.4-stretch

RUN dpkg --add-architecture ppc64el \
    && apt-get update && apt-get install -y \
    vim \
    git \
    curl \
    automake \
    libtool \
    help2man \
    libseccomp2 \
    libseccomp-dev \
    crossbuild-essential-ppc64el \
    libseccomp-dev:ppc64el\
    libfuse-dev:ppc64el \
    libpam-dev:ppc64el \
    lxcfs \
    btrfs-progs \
    --no-install-recommends \
    && apt-get clean

ADD . /go/src/github.com/alibaba/pouch
WORKDIR /go/src/github.com/alibaba/pouch

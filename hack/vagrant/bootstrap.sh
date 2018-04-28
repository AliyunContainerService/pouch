#!/usr/bin/env bash

set -euo pipefail

GOVERSION="${GOVERSION:-1.9.1}"
SOURCEPATH="${SOURCEPATH:-/go/src/github.com/alibaba/pouch}"


# install_pkgs install pkgs such as git/libncurses5-dev/... 
# which is useful for pouch development environment
install_pkgs(){
    apt-get update
    apt-get install -y git \
                libncurses5-dev  \
                libslang2-dev \
                gettext  \
                zlib1g-dev \
                libselinux1-dev \
                debhelper \
                lsb-release \
                pkg-config \
                po-debconf \
                autoconf \
                automake \
                autopoint \
                libtool
}

# install_go install go with version $GOVERSION with goenv
install_go(){
    git clone --depth=1 https://github.com/syndbg/goenv.git ~/.goenv

    PROFILE="$HOME/.bashrc"
    echo 'export GOENV_ROOT="$HOME/.goenv"' >> $PROFILE
    echo 'export PATH="$GOENV_ROOT/bin:$PATH"' >> $PROFILE
    echo 'eval "$(goenv init -)"' >> $PROFILE
    echo 'export GOPATH="/go"' >> $PROFILE

    $HOME/.goenv/bin/goenv install $GOVERSION
    $HOME/.goenv/bin/goenv global $GOVERSION
}

# install_docker install docker with get.docker.com's shell script
install_docker(){
    curl -fsSL get.docker.com -o get-docker.sh
    sh get-docker.sh
}

main() {
    install_pkgs
    install_go
    install_docker
}

main
#!/usr/bin/env bash

set -e

TMP=$(mktemp -d /tmp/pouch.XXXXXX)

MOUNTDIR=/root/rpm
PACKAGEDIR=/root/rpm/package/rpm
[ -d $PACKAGEDIR ] || mkdir -p $PACKAGEDIR

BASEDIR=/go/src/github.com/alibaba
SERVICEDIR=$BASEDIR/pouch/hack/package/rpm/service
SCRIPTSDIR=$BASEDIR/pouch/hack/package/rpm/scripts

POUCHDIR=$TMP/source
[ -d "$POUCHDIR" ] || mkdir -p "$POUCHDIR"
BINDIR=$POUCHDIR/bin
[ -d "$BINDIR" ] || mkdir -p "$BINDIR"
LIBDIR=$POUCHDIR/lib
[ -d "$LIBDIR" ] || mkdir -p "$LIBDIR"

USRBASE=/usr/local
USRBINDIR=$USRBASE/bin
USRLIBDIR=$USRBASE/lib

# ARCHITECTURE, The architecture name. Usually matches 'uname -m'.
ARCHITECTURE=$(uname -m)
LICENSE='Apache License 2.0'
# The category of Pouch
CATEGORY='Tools/Pouch'
# The maintainer of this package.
MAINTAINER='Pouch pouch-dev@list.alibaba-inc.com'
VENDOR='Pouch'

LIB_NVIDIA_VERSION="1.0.0-rc.2"
NVIDIA_RUNTIME_VERSION="1.4.0-1"

# build lxcfs and install containerd, runc and pouch
function build_pouch()
{
    pushd $BASEDIR/pouch
    # install containerd, runc and lxcfs dependencies for packaging
    make package-dependencies
    # build pouch
    make build
    make install DEST_DIR="$POUCHDIR"
    popd

    # make sure related binaries are included by pouch package
    cp $USRBINDIR/containerd* $USRBINDIR/ctr $USRBINDIR/runc "$BINDIR"

}

# install nvidia-container-runtime
function build_nvidia_runtime(){
    echo "Downloading libnvidia-container."
    wget --quiet "https://github.com/NVIDIA/libnvidia-container/releases/download/v${LIB_NVIDIA_VERSION}/libnvidia-container_${LIB_NVIDIA_VERSION}_x86_64.tar.xz" -P "${TMP}"
    tar -xf "${TMP}/libnvidia-container_${LIB_NVIDIA_VERSION}_x86_64.tar.xz" -C "${TMP}"
    cp "${TMP}/libnvidia-container_${LIB_NVIDIA_VERSION}/usr/local/bin/nvidia-container-cli" "${BINDIR}/"
    cp "${TMP}/libnvidia-container_${LIB_NVIDIA_VERSION}/usr/local/lib/libnvidia-container.so" "${LIBDIR}/"
    cp "${TMP}/libnvidia-container_${LIB_NVIDIA_VERSION}/usr/local/lib/libnvidia-container.so.1" "${LIBDIR}/"
    cp "${TMP}/libnvidia-container_${LIB_NVIDIA_VERSION}/usr/local/lib/libnvidia-container.so.1.0.0" "${LIBDIR}/"

    echo "Downloading nvidia-container-runtime."
    wget --quiet "https://github.com/NVIDIA/nvidia-container-runtime/archive/v${NVIDIA_RUNTIME_VERSION}.tar.gz" -P "${TMP}"
    mkdir -p "${GOPATH}/src/github.com/NVIDIA"
    tar -xzf "${TMP}/v${NVIDIA_RUNTIME_VERSION}.tar.gz" -C "${GOPATH}/src/github.com/NVIDIA"
    mv "${GOPATH}/src/github.com/NVIDIA/nvidia-container-runtime-${NVIDIA_RUNTIME_VERSION}" "${GOPATH}/src/github.com/NVIDIA/nvidia-container-runtime"
    go build -o "${BINDIR}/nvidia-container-runtime-hook" "github.com/NVIDIA/nvidia-container-runtime/hook/nvidia-container-runtime-hook"
}

function build_rpm ()
{
    pushd $MOUNTDIR
    # import gpg keys
    gpg --import $MOUNTDIR/keys/gpg
    gpg --import $MOUNTDIR/keys/secretkey
    rpm --import $MOUNTDIR/keys/gpg
    popd

    # configure gpg
    echo "%_gpg_name Pouch Release" >> /root/.rpmmacros

    fpm -f -s dir \
          -t rpm \
          -n pouch \
          -v "$VERSION" \
          --iteration "$ITERATION" \
          -a "$ARCHITECTURE" \
          -p "$PACKAGEDIR" \
          --description 'Pouch is an open-source project created by Alibaba Group to promote the container technology movement.

    Pouch'"'"'s vision is to advance container ecosystem and promote container standards OCI, so that container technologies become the foundation for application development in the Cloud era.

    Pouch can pack, deliver and run any application. It provides applications with a lightweight runtime environment with strong isolation and minimal overhead. Pouch isolates applications from varying runtime environment, and minimizes operational workload. Pouch minimizes the effort for application developers to write Cloud-native applications, or to migrate legacy ones to a Cloud platform.' \
          --url 'https://github.com/alibaba/pouch' \
          --before-install $SCRIPTSDIR/before-install.sh \
          --after-install $SCRIPTSDIR/after-install.sh \
          --before-remove $SCRIPTSDIR/before-remove.sh \
          --after-remove $SCRIPTSDIR/after-remove.sh \
          --rpm-posttrans $SCRIPTSDIR/after-trans.sh \
          --license "$LICENSE" \
          --verbose \
          --category "$CATEGORY" \
          -m "$MAINTAINER" \
          --vendor "$VENDOR" \
          --rpm-sign \
          -d pam-devel \
          -d fuse-devel \
          -d fuse-libs \
          -d fuse \
          "$BINDIR/"=/usr/local/bin/ \
          "$LIBDIR/"=/usr/lib64/ \
          "$SERVICEDIR/"=/usr/lib/systemd/system/ \
          "$USRBINDIR/lxcfs"=/usr/bin/pouch-lxcfs \
          "$USRLIBDIR/lxcfs/liblxcfs.so"=/usr/lib64/libpouchlxcfs.so \

}

function main()
{
	echo "Building rpm package."
	build_pouch
	build_nvidia_runtime
	build_rpm
}

main "$@"

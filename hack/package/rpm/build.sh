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
LXC_DIR=$TMP/lxc
[ -d "$LXC_DIR" ] || mkdir -p "$LXC_DIR"

# Dependency
# lxcfs stable branch
LXC_BRANCH="stable-2.0"
CONTAINERD_VERSION="1.0.3"
RUNC_VERSION="1.0.0-rc4-1"

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

# build lxcfs
function build_lxcfs ()
{
    pushd "$LXC_DIR"
    git clone -b "$LXC_BRANCH" https://github.com/lxc/lxcfs.git && cd lxcfs
    
    # change liblxcfs.so to libpouchlxcfs.so
    grep -l -r "liblxcfs" . | xargs sed -i 's/liblxcfs/libpouchlxcfs/g'
    ./bootstrap.sh > /dev/null 2>&1
    ./configure > /dev/null 2>&1
    make install DESTDIR="$LXC_DIR" > /dev/null 2>&1
    popd
}

# install containerd, runc and pouch
function build_pouch()
{
    # install containerd
    echo "Downloading containerd."
    wget --quiet "https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}.linux-amd64.tar.gz" -P "$TMP"
    tar xf "$TMP/containerd-${CONTAINERD_VERSION}.linux-amd64.tar.gz" -C "$TMP" && cp -f "$TMP"/bin/* "$BINDIR/"

    # install runc
    echo "Downloading runc."
    wget --quiet "https://github.com/alibaba/runc/releases/download/v${RUNC_VERSION}/runc.amd64" -P "$BINDIR/"
    chmod +x "$BINDIR/runc.amd64"
    mv "$BINDIR/runc.amd64" "$BINDIR/runc"

    # build pouch
    echo "Building pouch."
    pushd $BASEDIR/pouch
    make install DESTDIR="$POUCHDIR"
    popd
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
          "$LXC_DIR/usr/local/bin/lxcfs"=/usr/bin/pouch-lxcfs \
          "$LXC_DIR/usr/local/lib/lxcfs/libpouchlxcfs.so"=/usr/lib64/libpouchlxcfs.so \

}

function main()
{
	echo "Building rpm package."
	build_pouch
	build_lxcfs
	build_nvidia_runtime
	build_rpm
}

main "$@"

#!/usr/bin/env sh

set -e
# This script is to build pouch rpm package as follows,
# Following the below command to build rpm
# 1. Build pouch:rpm image
#	cd hack/package/rpm
#	docker build -t pouch:rpm .
# 2. Mount a directory which contains gpg keys, eg
#    $ tree /root/rpm/
#		rpm
#		├── config
#		├── keys
#		│   ├── gpg
#		│   └── secretkey
#
# Note:
# In the config file you should configure the version, iteration, et.al
#
# VERSION, the version to give to the package, eg:
# VERSION='0.1.0'
#
# The iteration to give to the package. RPM calls this the 'release'.
# FreeBSD calls it 'PORTREVISION'. Debian calls this 'debian_revision', eg:
# ITERATION='1.el7.centos'
#
# ARCHITECTURE, The architecture name. Usually matches 'uname -m'.
# ARCHITECTURE='x86_64'
#
# the branch to build pouch
# POUCH_BRANCH='0.1.x'
# POUCH_COMMIT='6be2080cd9837e9b8a0039c2d21521bb00a30c84'
#
# lxcfs stable branch
# LXC_TAG='stable-2.0'
# LXC_DIR=$TMP/lxc
#
# 3. Run the following command, and enter your pass phrase to sign rpm package
# docker run -it -v /root/rpm/:/root/rpm pouch:rpm bash -c hack/package/build.sh
#
# 4. In this example rpm package will be output in '/root/rpm/package/' directory

DIR="$( cd "$( dirname "$0" )" && pwd )"

TMP=$(mktemp -d /tmp/pouch.XXXXXX)

MOUNTDIR=/root/rpm
PACKAGEDIR=/root/rpm/package

BASEDIR=/go/src/github.com/alibaba
SERVICEDIR=$DIR/rpm/service
SCRIPTSDIR=$DIR/rpm/scripts

POUCHDIR=$TMP/source
[ -d $POUCHDIR ] || mkdir -p $POUCHDIR
BINDIR=$POUCHDIR/bin
[ -d $BINDIR ] || mkdir -p $BINDIR

SUMMARY='The open-source reliable application container engine.'

# load config info
source $MOUNTDIR/config

# build lxcfs
function build_lxcfs ()
{
    mkdir -p $LXC_DIR && pushd $LXC_DIR
    git clone -b $LXC_TAG https://github.com/lxc/lxcfs.git && cd lxcfs
    ./bootstrap.sh > /dev/null 2>&1
    ./configure > /dev/null 2>&1
    make install DESTDIR=$LXC_DIR > /dev/null 2>&1
    popd
}

# install containerd, runc and pouch
function build_pouch()
{
    # install containerd
    echo "Downloading containerd."
    wget --quiet https://github.com/containerd/containerd/releases/download/v1.0.0/containerd-1.0.0.linux-amd64.tar.gz -P $TMP
    tar xf $TMP/containerd-1.0.0.linux-amd64.tar.gz -C $TMP && cp -f $TMP/bin/* $BINDIR/

    # install runc
    echo "Downloading runc."
    wget --quiet https://github.com/opencontainers/runc/releases/download/v1.0.0-rc4/runc.amd64 -P $BINDIR/
    chmod +x $BINDIR/runc.amd64
    mv $BINDIR/runc.amd64 $BINDIR/runc

    # build pouch
    echo "Building pouch."
    pushd $BASEDIR/pouch
    git fetch && git checkout $POUCH_BRANCH && git checkout -q $POUCH_COMMIT
    make install DESTDIR=$POUCHDIR
    popd
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
    echo "%_gpg_name Pouch Packages RPM Signing Key" >> /root/.rpmmacros

    fpm -f -s dir \
          -t rpm \
          -n pouch \
          -v $VERSION \
          --iteration $ITERATION \
          -a $ARCHITECTURE \
          -p $PACKAGEDIR \
          --description 'Pouch is an open-source project created by Alibaba Group to promote the container technology movement.

    Pouchs vision is to advance container ecosystem and promote container standards OCI, so that container technologies become the foundation for application development in the Cloud era.

    Pouch can pack, deliver and run any application. It provides applications with a lightweight runtime environment with strong isolation and minimal overhead. Pouch isolates applications from varying runtime environment, and minimizes operational workload. Pouch minimizes the effort for application developers to write Cloud-native applications, or to migrate legacy ones to a Cloud platform.' \
          --url 'https://github.com/alibaba/pouch' \
          --before-install $SCRIPTSDIR/before-install.sh \
          --after-install $SCRIPTSDIR/after-install.sh \
          --before-remove $SCRIPTSDIR/before-remove.sh \
          --after-remove $SCRIPTSDIR/after-remove.sh \
          --rpm-posttrans $SCRIPTSDIR/after-trans.sh \
          --license 'Apache License 2.0' \
          --verbose \
          --category 'Tools/Pouch' \
          -m 'Pouch pouch-dev@list.alibaba-inc.com' \
          --vendor Pouch \
          --rpm-sign \
          -d pam-devel \
          -d fuse-devel \
          -d fuse-libs \
          $BINDIR/=/usr/local/bin/ \
          $SERVICEDIR/=/usr/lib/systemd/system/ \
          $LXC_DIR/usr/local/bin/lxcfs=/usr/bin/lxcfs \
          $LXC_DIR/usr/local/lib/lxcfs/liblxcfs.so=/usr/lib64/liblxcfs.so \
          $LXC_DIR/usr/local/share/=/usr/share

}

function main()
{
	echo "Building rpm package."
	build_pouch
	build_lxcfs
	build_rpm

	# echo "Building deb package."
	# echo "TODO: build deb"
}

main "$@"

#!/bin/bash

#
# Install lxcfs from source code in CentOS
#

yes | yum install autotools-dev m4 autoconf2.13 autobook autoconf-archive gnu-standards autoconf-doc libtool
yes | yum install "fuse-devel.$(uname -p)"
yes | yum install "pam-devel.$(uname -p)"
yes | yum install "fuse.$(uname -p)"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

cd "$TMP" &&
git clone -b stable-2.0 https://github.com/lxc/lxcfs.git &&
cd lxcfs

./bootstrap.sh
./configure 
make install

ln -s /usr/local/bin/lxcfs /usr/bin/lxcfs
mkdir -p /var/lib/lxcfs

which lxcfs
exit $?

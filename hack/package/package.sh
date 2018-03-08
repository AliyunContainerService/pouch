#!/usr/bin/env bash

set -e

# VERSION the version to give to the package
VERSION=${1:?"VERSION is required."}

# The iteration to give to the package. RPM calls this the 'release'.
# FreeBSD calls it 'PORTREVISION'. Debian calls this 'debian_revision', eg:
ITERATION=${2:?"ITERATION is required"}

# the gpg keys directory
KEYDIR=${3:?"GPG key is required"}

# By default we will build deb and rpm packages
PKGTYPE=$4

# The time to build packages
BUILDTIME="$(date --rfc-2822)"

# The commit ID used to build package
GITCOMMIT=$( git rev-parse HEAD )

DIR="$( cd "$( dirname "$0" )/../.." && pwd )"

# Replace Version, BuildTime and CommitID in version.go
pushd $DIR
sed -i "s#^const Version.*#const Version = \"$VERSION\"#g" version/version.go
sed -i "s#^var BuildTime.*#var BuildTime = \"$BUILDTIME\"#g" version/version.go
sed -i "s#^var GitCommit.*#var GitCommit = \"$GITCOMMIT\"#g" version/version.go

# build rpm packages
function build_rpm() {
	# build images
	docker build --network host -t pouch:rpm -f $DIR/hack/package/rpm/centos-7/Dockerfile.x86_64 .
	(( $? != 0 )) && echo "failed to build pouch:rpm image." && exit 1

	docker run --network host -it --rm \
		-e VERSION="$VERSION" \
		-e ITERATION="$ITERATION" \
		-v $KEYDIR:/root/rpm \
		pouch:rpm
}

# build deb packages
function build_deb() {
	# build images
	docker build --network host -t pouch:deb -f $DIR/hack/package/deb/ubuntu-xenial/Dockerfile.x86_64 .
	(( $? != 0 )) && echo "failed to build pouch:deb image." && exit 1

	docker run --network host -it --rm \
		-e VERSION="$VERSION" \
		-v $KEYDIR/:/root/deb \
		pouch:deb
}

function main() {
	if [[ $PKGTYPE == "rpm" ]]; then
		build_rpm
	elif [[ $PKGTYPE == "deb" ]]; then
		build_deb
	else
		build_rpm
		build_deb
	fi
}

main


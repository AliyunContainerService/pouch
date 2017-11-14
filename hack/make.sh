#!/usr/bin/env bash
set -e

# This script is to build pouch binaries and execute pouch tests.

TMP=`mktemp -d /tmp/pouch.XXXXXX`

DIR="$( cd "$( dirname "$0" )" && pwd )"
cd $DIR/..

# This function would build pouch binaries and execute unit tests.
function build_test ()
{
	# create a image which contain pouch dev and test environment.
	docker build --quiet -t pouch:test .

	# build pouch binaries and execute unit tests.
	echo "Build pouch binaries and execute unit tests."
	docker run -ti -v `pwd`:/go/src/github.com/alibaba/pouch pouch:test bash -c "make check && make build && make unit-test"
}

# install pouch and essential binaries: containerd, runc and so on
function install_pouch ()
{
	# install containerd
	echo "Download and install containerd."
	wget --quiet https://github.com/containerd/containerd/releases/download/v1.0.0-beta.3/containerd-1.0.0-beta.3.linux-amd64.tar.gz -P $TMP
	tar xf $TMP/containerd-1.0.0-beta.3.linux-amd64.tar.gz -C $TMP && cp -f $TMP/bin/* /usr/local/bin/

	# install runc
	echo "Download and install runc."
	wget --quiet https://github.com/opencontainers/runc/releases/download/v1.0.0-rc4/runc.amd64 -P /usr/local/bin
	chmod +x /usr/local/bin/runc.amd64
	mv /usr/local/bin/runc.amd64 /usr/local/bin/runc

	# copy pouch daemon and pouch cli to PATH
	echo "Install pouch."
	cp -f pouch pouchd /usr/local/bin/
}

function main ()
{
	build_test
	install_pouch

	#start pouch daemon
	echo "start pouch daemon"
	pouchd > $TMP/log 2>&1 &

	# wait until pouch daemon is ready
	while true;
	do
		if [ -S /var/run/pouchd.sock ];then
			break
		else
			sleep 1
		fi
	done

	echo "verify pouch version"
	pouch version

	# This scripts can't get environment variables in travis ci.
	PATH=$PATH:/home/travis/.gimme/versions/go1.8.3.linux.amd64/bin/
	cd $DIR/../test
	go test
}

main "$@"

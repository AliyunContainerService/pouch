#!/usr/bin/env bash
set -e

# This script is to build pouch binaries and execute pouch tests.

TMP=$(mktemp -d /tmp/pouch.XXXXXX)

DIR="$( cd "$( dirname "$0" )/.." && pwd )"
cd $DIR/
SOURCEDIR=/go/src/github.com/alibaba/pouch
IMAGE=pouch:test

# install pouch and essential binaries: containerd, runc and so on
function install_pouch ()
{
	# install containerd
	echo "Download and install containerd."
	wget --quiet https://github.com/containerd/containerd/releases/download/v1.0.0/containerd-1.0.0.linux-amd64.tar.gz -P $TMP
	tar xf $TMP/containerd-1.0.0.linux-amd64.tar.gz -C $TMP && cp -f $TMP/bin/* /usr/local/bin/

	# install runc
	echo "Download and install runc."
	wget --quiet https://github.com/opencontainers/runc/releases/download/v1.0.0-rc4/runc.amd64 -P /usr/local/bin
	chmod +x /usr/local/bin/runc.amd64
	mv /usr/local/bin/runc.amd64 /usr/local/bin/runc

	# copy pouch daemon and pouch cli to PATH
	echo "Install pouch."
	cp -f $DIR/pouch $DIR/pouchd /usr/local/bin/
}

function target()
{
	case $1 in
	check)
		docker run -ti -v $(pwd):$SOURCEDIR $IMAGE bash -c "make check"
		;;
	build)
		docker run -ti -v $(pwd):$SOURCEDIR $IMAGE bash -c "make build"
		install_pouch
		;;
	unit-test)
		docker run -ti -v $(pwd):$SOURCEDIR $IMAGE bash -c "make unit-test"
		;;
	integration-test)
		docker run -ti -v $(pwd):$SOURCEDIR $IMAGE bash -c "cd test && go test -c -o integration-test"

		if [[ $SOURCEDIR != $DIR ]];then
			[ -d $SOURCEDIR ] && rm -rf $SOURCEDIR
			POUCHTOPDIR=$(dirname $SOURCEDIR)
			[ ! -d $POUCHTOPDIR ] && mkdir -p $POUCHTOPDIR
			ln -sf $DIR/ $SOURCEDIR
		fi

		#start pouch daemon
		echo "start pouch daemon"
		pouchd > $TMP/log 2>&1 &

		# wait until pouch daemon is ready
		daemon_timeout_time=30
		while true;
		do
			if [[ -S /var/run/pouchd.sock ]];then
				echo "Succeed to start pouch daemon"
				break
			elif (( $((daemon_timeout_time--)) == 0 ));then
				echo "Failed to start pouch daemon"
				return 1
			else
				sleep 1
			fi
		done

		pouch pull registry.hub.docker.com/library/busybox:latest >/dev/null

		echo "verify pouch version"
		pouch version

		# If test is failed, print pouch daemon log.
		$DIR/test/integration-test || { echo "pouch daemon log:"; cat $TMP/log; return 1; } 
		;;
	*)
		echo "no such target: $target"
		return 1
		;;
	esac
}


function main ()
{
	docker build --quiet -t $IMAGE .

	if [[ $# < 1 ]]; then
		targets="check build unit-test integration-test"
	else
		targets=($@)
	fi

	for target in ${targets[@]}; do
		target $target
	done
}

main "$@"

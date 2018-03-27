#!/usr/bin/env bash
set -ex

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
	wget --quiet https://github.com/alibaba/runc/releases/download/v1.0.0-rc4-1/runc.amd64 -P /usr/local/bin
	chmod +x /usr/local/bin/runc.amd64
	mv /usr/local/bin/runc.amd64 /usr/local/bin/runc

	# copy pouch daemon and pouch cli to PATH
	echo "Install pouch."
	cp -f $DIR/pouch $DIR/pouchd /usr/local/bin/
	
	# install lxcfs
	echo "Install lxcfs"
	if grep -qi "ubuntu" /etc/issue ; then
		apt-get install -y lxcfs
		if (( $? != 0 )); then
			add-apt-repository ppa:ubuntu-lxc/lxcfs-stable -y
			apt-get update
			apt-get install -y lxcfs
		fi
	else
		sh -x $DIR/hack/install_lxcfs_on_centos.sh
	fi

	# install nsenter
	echo "Install nsenter"
	if grep -qi "ubuntu" /etc/issue ; then
		apt-get install libncurses5-dev libslang2-dev gettext zlib1g-dev libselinux1-dev debhelper lsb-release pkg-config po-debconf autoconf automake autopoint libtool
		wget https://www.kernel.org/pub/linux/utils/util-linux/v2.24/util-linux-2.24.1.tar.gz -P $TMP
		tar xf $TMP/util-linux-2.24.1.tar.gz -C $TMP && cd $TMP/util-linux-2.24.1
		./autogen.sh
		autoreconf -vfi
		./configure && make 
		cp ./nsenter /usr/local/bin
		cd $DIR/
	else
		yum install -y util-linux
	fi
}

# Install dumb-init by downloading the binary.
function install_dumb_init
{
    wget -O /usr/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.1/dumb-init_1.2.1_amd64 || return 1
    chmod +x /usr/bin/dumb-init
}

function target()
{
	case $1 in
	check)
		docker run --rm -v $(pwd):$SOURCEDIR $IMAGE bash -c "make check"
		;;
	build)
		docker run --rm -v $(pwd):$SOURCEDIR $IMAGE bash -c "make build"  >$TMP/build.log ||
		    { echo "make build log:"; cat $TMP/build.log; return 1; }
		install_pouch  >$TMP/install.log ||
		    { echo "install pouch log:"; cat $TMP/install.log; return 1; }
		;;
	unit-test)
		docker run --rm -v $(pwd):$SOURCEDIR $IMAGE bash -c "make unit-test"
		;;
	cri-test)
		cd $SOURCEDIR
		env PATH=$GOROOT/bin:$PATH $SOURCEDIR/hack/cri-test/test-cri.sh
		;;
	integration-test)

	    install_dumb_init || echo "Warning: dumb-init install failed! rich container related tests will be skipped"
		docker run --rm -v $(pwd):$SOURCEDIR $IMAGE bash -c "cd test && go test -c -o integration-test"

		if [[ $SOURCEDIR != $DIR ]];then
			[ -d $SOURCEDIR ] && rm -rf $SOURCEDIR
			POUCHTOPDIR=$(dirname $SOURCEDIR)
			[ ! -d $POUCHTOPDIR ] && mkdir -p $POUCHTOPDIR
			ln -sf $DIR/ $SOURCEDIR
		fi

		#start pouch daemon
		echo "start pouch daemon"
		if stat /usr/bin/lxcfs ; then
			pouchd --enable-lxcfs=true --lxcfs=/usr/bin/lxcfs > $TMP/log 2>&1 &
		else
			pouchd > $TMP/log 2>&1 &
		fi

		# wait until pouch daemon is ready
		daemon_timeout_time=30
		while true;
		do
			if pouch version; then
				echo "Succeed to start pouch daemon"
				break
			elif (( $((daemon_timeout_time--)) == 0 ));then
				echo "Failed to start pouch daemon"
				echo "pouch daemon log:"
				cat $TMP/log 
				return 1
			else
				sleep 1
			fi
		done

		pouch pull registry.hub.docker.com/library/busybox:latest >/dev/null

		echo "verify pouch version"
		pouch version

		# If test is failed, print pouch daemon log.
		$DIR/test/integration-test -test.v -check.v || { echo "pouch daemon log:"; cat $TMP/log; return 1; } 
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

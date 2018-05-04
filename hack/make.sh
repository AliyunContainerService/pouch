#!/usr/bin/env bash
set -ex

# This script is to build pouch binaries and execute pouch tests.

TMP=$(mktemp -d /tmp/pouch.XXXXXX)
CONTAINERD_VERSION=
RUNC_VERSION=
NSENTER_VERSION=
DUMB_INIT_VERSION=

DIR="$( cd "$( dirname "$0" )/.." && pwd )"
cd $DIR/
SOURCEDIR=/go/src/github.com/alibaba/pouch
IMAGE=pouch:test
if [[ $SOURCEDIR != $DIR ]];then
	[ -d $SOURCEDIR ] && rm -rf $SOURCEDIR
	POUCHTOPDIR=$(dirname $SOURCEDIR)
	[ ! -d $POUCHTOPDIR ] && mkdir -p $POUCHTOPDIR
	ln -sf $DIR/ $SOURCEDIR
fi

function get_containerd_version()
{
	if which containerd &>/dev/null; then
		CONTAINERD_VERSION=$(containerd -v|cut -d " " -f 3)
	fi
}

function get_runc_version()
{
	if which runc &>/dev/null; then
		RUNC_VERSION=$(runc -v|head -1| cut -d " " -f 3)
	fi
}

function get_nsenter_version()
{
	if which nsenter &>/dev/null; then
		NSENTER_VERSION=$(nsenter -V | cut -d " " -f 4)
	fi	
}

function get_dumb_init_version()
{
	if which dumb-init &>/dev/null; then
		DUMB_INIT_VERSION=$(dumb-init -V 2>&1 | cut -d " " -f 2)
	fi	
}

function install_containerd()
{
	echo "Try installing containerd"
	get_containerd_version
	if [[ "$CONTAINERD_VERSION" == "v1.0.3" ]]; then
		echo "Containerd already installed."
	else
		echo "Download and install containerd."
		wget --quiet https://github.com/containerd/containerd/releases/download/v1.0.3/containerd-1.0.3.linux-amd64.tar.gz -P $TMP
		tar xf $TMP/containerd-1.0.3.linux-amd64.tar.gz -C $TMP && cp -f $TMP/bin/* /usr/local/bin/
	fi;
}

function install_runc()
{
	echo "Try installing runc"
	get_runc_version
	if [[ "$RUNC_VERSION" == "1.0.0-rc4-1" ]]; then
		echo "Runc already installed."
	else
		echo "Download and install runc."
		wget --quiet https://github.com/alibaba/runc/releases/download/v1.0.0-rc4-1/runc.amd64 -P /usr/local/bin
		chmod +x /usr/local/bin/runc.amd64
		mv /usr/local/bin/runc.amd64 /usr/local/bin/runc
	fi;
}

function install_lxcfs()
{
	echo "Try installing lxcfs"
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
}

function install_nsenter()
{
	echo "Try installing nsenter"
	get_nsenter_version
	if grep -qi "ubuntu" /etc/issue ; then
		if [[ "$NSENTER_VERSION" == "2.24.1" ]]; then
			echo "Nsenter already installed."
		else
			echo "Download and install nsenter."
			apt-get -y install libncurses5-dev libslang2-dev gettext zlib1g-dev libselinux1-dev debhelper lsb-release pkg-config po-debconf autoconf automake autopoint libtool
			wget --quiet https://www.kernel.org/pub/linux/utils/util-linux/v2.24/util-linux-2.24.1.tar.gz -P $TMP
			tar xf $TMP/util-linux-2.24.1.tar.gz -C $TMP && cd $TMP/util-linux-2.24.1
			./autogen.sh
			autoreconf -vfi
			./configure && make 
			cp ./nsenter /usr/local/bin
			cd $DIR/
		fi
	else
		yum install -y util-linux
	fi

}

# Install dumb-init by downloading the binary.
function install_dumb_init
{
	echo "Try installing dumb-init"
	get_dumb_init_version
	if [[ "$DUMB_INIT_VERSION" == "v1.2.1" ]]; then
		echo "Dumb-init already installed."
	else
		echo "Download and install dumb-init."
		wget --quiet -O /tmp/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.1/dumb-init_1.2.1_amd64
		mv tmp/dumb-init /usr/bin/
		chmod +x /usr/bin/dumb-init
	fi
}

# install pouch and essential binaries: containerd, runc and so on
function install_pouch ()
{
	# install containerd
	install_containerd
	# install runc
	install_runc
	# copy pouch daemon and pouch cli to PATH
	echo "Install pouch."
	cp -f $DIR/pouch $DIR/pouchd /usr/local/bin/
	# install lxcfs
	install_lxcfs
	# install nsenter
	install_nsenter
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
		docker run --rm -v $(pwd):$SOURCEDIR -e GOPATH=/go:$SOURCEDIR/extra/libnetwork/Godeps/_workspace $IMAGE bash -c "cd test && go test -c -o integration-test"

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

		pouch pull registry.hub.docker.com/library/busybox:1.28 >/dev/null

		echo "verify pouch version"
		pouch version

        # copy tls file
        cp -rf $DIR/test/tls /tmp/

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

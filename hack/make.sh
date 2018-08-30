#!/usr/bin/env bash
set -ex

#
# This script is used to build pouch binaries and execute pouch tests.
#
TMP=$(mktemp -d /tmp/pouch.XXXXXX)
DIR="$( cd "$( dirname "$0" )/.." && pwd )"
cd "$DIR/"

#
# Get the version of dependencies from corresponding commands and compare them with the required version.
# If they don't meet the requirement, this script will install them.
#
CONTAINERD_VERSION=
REQUIRED_CONTAINERD_VERSION="1.0.3"
RUNC_VERSION=
REQUIRED_RUNC_VERSION="1.0.0-rc4-1"
NSENTER_VERSION=
REQUIRED_NSENTER_VERSION="2.24.1"
DUMB_INIT_VERSION=
REQUIRED_DUMB_INIT_VERSION="1.2.1"

SOURCEDIR=/go/src/github.com/alibaba/pouch

IMAGE="registry.hub.docker.com/letty/pouchci:latest"
if [[ $SOURCEDIR != $DIR ]];then
	[ -d $SOURCEDIR ] && rm -rf $SOURCEDIR
	POUCHTOPDIR=$(dirname $SOURCEDIR)
	[ ! -d "$POUCHTOPDIR" ] && mkdir -p "$POUCHTOPDIR"
	ln -sf "$DIR/" $SOURCEDIR
fi

#
# CAL_INTEGRATION_TEST_COVERAGE indicates whehter or not calculate integration test coverage.
# By default it is yes.
#
CAL_INTEGRATION_TEST_COVERAGE=${CAL_INTEGRATION_TEST_COVERAGE:-"yes"}
if [[ $CAL_INTEGRATION_TEST_COVERAGE == "yes" ]]; then
	POUCHD="pouchd-test -test.coverprofile=$DIR/integrationcover.out DEVEL"
else
	POUCHD="pouchd"
fi

function get_containerd_version
{
	if which containerd &>/dev/null; then
		CONTAINERD_VERSION=$(containerd -v|cut -d " " -f 3|cut -c 2-)
	fi
}

function get_runc_version
{
	if which runc &>/dev/null; then
		RUNC_VERSION=$(runc -v|head -1| cut -d " " -f 3)
	fi
}

function get_nsenter_version
{
	if which nsenter &>/dev/null; then
		NSENTER_VERSION=$(nsenter -V | cut -d " " -f 4)
	fi	
}

function get_dumb_init_version
{
	if which dumb-init &>/dev/null; then
		DUMB_INIT_VERSION=$(dumb-init -V 2>&1 | cut -d " " -f 2|cut -c 2-)
	fi	
}

function install_containerd
{
	echo "Try installing containerd"
	get_containerd_version
	if [[ "$CONTAINERD_VERSION" == "$REQUIRED_CONTAINERD_VERSION" ]]; then
		echo "Containerd already installed."
	else
		echo "Download and install containerd."
		wget --quiet \
			"https://github.com/containerd/containerd/releases/download/v${REQUIRED_CONTAINERD_VERSION}/containerd-${REQUIRED_CONTAINERD_VERSION}.linux-amd64.tar.gz" -P "$TMP"
		tar xf "$TMP/containerd-${REQUIRED_CONTAINERD_VERSION}.linux-amd64.tar.gz" -C "$TMP" &&
			cp -f "$TMP"/bin/* /usr/local/bin/
	fi;
}

function install_runc
{
	echo "Try installing runc"
	get_runc_version
	if [[ "$RUNC_VERSION" == "$REQUIRED_RUNC_VERSION" ]]; then
		echo "Runc already installed."
	else
		echo "Download and install runc."
		wget --quiet \
			"https://github.com/alibaba/runc/releases/download/v${REQUIRED_RUNC_VERSION}/runc.amd64" -P /usr/local/bin
		chmod +x /usr/local/bin/runc.amd64
		mv /usr/local/bin/runc.amd64 /usr/local/bin/runc
	fi;
}

function install_lxcfs
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
		sh -x "$DIR/hack/install_lxcfs_on_centos.sh"
	fi
}

function install_criu
{
	echo "Try installing criu"
	if grep -qi "ubuntu" /etc/issue ; then
		apt-get update
		apt-get install -y criu
	fi
}

# local-persist is a volume plugin
function install_local_persist
{
	echo "Try installing local-persist"
	wget --quiet -O /tmp/local-persist \
		https://github.com/CWSpear/local-persist/releases/download/v1.3.0/local-persist-linux-amd64
	chmod +x /tmp/local-persist
	mv /tmp/local-persist /usr/local/bin/
}

# clean the local-persist
function clean_local_persist
{
	echo "Try cleaning local-persist"
	pid=$(pgrep local-persist)

	if [[ $pid ]]; then
		echo "Try killing local-persist process"
		kill -9 "$pid"
	fi

	echo "Try removing local-persist.sock"
	rm -rf /var/run/docker/plugins/local-persist.sock
}

function install_nsenter
{
	echo "Try installing nsenter"
	get_nsenter_version
	if grep -qi "ubuntu" /etc/issue ; then
		if [[ "$NSENTER_VERSION" == "$REQUIRED_NSENTER_VERSION" ]]; then
			echo "Nsenter already installed."
		else
			echo "Download and install nsenter."
			apt-get -y install \
				libncurses5-dev \
				libslang2-dev \
				gettext \
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
			wget --quiet \
				"https://www.kernel.org/pub/linux/utils/util-linux/v2.24/util-linux-${REQUIRED_NSENTER_VERSION}.tar.gz" -P "$TMP"
			tar xf "$TMP/util-linux-${REQUIRED_NSENTER_VERSION}.tar.gz" -C "$TMP" &&
				cd "$TMP/util-linux-${REQUIRED_NSENTER_VERSION}"
			./autogen.sh
			autoreconf -vfi
			./configure && make 
			cp ./nsenter /usr/local/bin
			cd "$DIR/"
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
	if [[ "$DUMB_INIT_VERSION" == "$REQUIRED_DUMB_INIT_VERSION" ]]; then
		echo "Dumb-init already installed."
	else
		echo "Download and install dumb-init."
		wget --quiet -O /tmp/dumb-init \
			 "https://github.com/Yelp/dumb-init/releases/download/v${REQUIRED_DUMB_INIT_VERSION}/dumb-init_${REQUIRED_DUMB_INIT_VERSION}_amd64"
		mv /tmp/dumb-init /usr/bin/
		chmod +x /usr/bin/dumb-init
	fi
}

# install pouch and essential binaries: containerd, runc and so on
function install_pouch 
{
	install_containerd
	install_runc
	# copy pouch daemon and pouch cli to PATH
	echo "Install pouch."
	if [[ $CAL_INTEGRATION_TEST_COVERAGE == "yes" ]]; then
		cp -f "$DIR/pouchd-test" /usr/local/bin/
	fi
	cp -f "$DIR/pouch" "$DIR/pouchd" /usr/local/bin/
	install_lxcfs
	install_nsenter
	install_criu
}

function target
{
	case $1 in
	check)
		docker run --rm -v "$(pwd):$SOURCEDIR" "$IMAGE" bash -c "make check"
		;;
	build)
		#
		# Also build pouchd-test binary if CAL_INTEGRATION_TEST_COVERAGE doesn't
		# equal to 'no'.
		#
		if [[ $CAL_INTEGRATION_TEST_COVERAGE == "yes" ]]; then
			docker run --rm -v "$(pwd):$SOURCEDIR" "$IMAGE" \
				bash -c "make testserver"  >"$TMP/build.log" ||
				{ echo "make build log:"; cat "$TMP/build.log"; return 1; }
		fi
		docker run --rm -v "$(pwd):$SOURCEDIR" "$IMAGE" \
			bash -c "make build"  >"$TMP/build.log" ||
			{ echo "make build log:"; cat "$TMP/build.log"; return 1; }

		install_pouch  >"$TMP/install.log" ||
			{ echo "install pouch log:"; cat "$TMP/install.log"; return 1; }
		;;
	unit-test)
		docker run --rm -v "$(pwd):$SOURCEDIR" "$IMAGE" \
			bash -c "make unit-test"
		;;
	cri-test)
		cd $SOURCEDIR
		env PATH="$GOROOT/bin:$PATH" "$SOURCEDIR/hack/cri-test/test-cri.sh"
		;;
	integration-test)
		
		install_dumb_init ||
			echo "Warning: dumb-init install failed!\
				 rich container related tests will be skipped"
	
		docker run --rm -v "$(pwd):$SOURCEDIR" \
			-e GOPATH=/go \
			$IMAGE \
			bash -c "cd test && go test -c -o integration-test"

		install_local_persist

        	# start local-persist
		echo "start local-persist volume plugin"
		local-persist > "$TMP/volume.log" 2 >&1 &

		# start pouch daemon
		echo "start pouch daemon"
		if stat /usr/bin/lxcfs ; then
			$POUCHD --debug --enable-lxcfs=true --add-runtime runv=runv \
				--lxcfs=/usr/bin/lxcfs > "$TMP/log" 2>&1 &
		else
			$POUCHD --debug --add-runtime runv=runv > "$TMP/log" 2>&1 &
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
				cat "$TMP/log"
				return 1
			else
				sleep 1
			fi
		done

		echo "verify pouch version"
		pouch version

		# copy tls file
		cp -rf "$DIR/test/tls" /tmp/

		# If test is failed, print pouch daemon log.
		set +e
		"$DIR/test/integration-test" -test.v -check.v

		if (( $? != 0 )); then
			echo "pouch daemon log:"
			cat "$TMP/log"
			clean_local_persist
			return 1
		fi

		clean_local_persist

		set -e
		;;
	*)
		echo "no such target: $target"
		return 1
		;;
	esac
}


function main 
{
	docker pull $IMAGE
	if (( $? != 0 )); then
		echo "ERR: pull $IMAGE failed"
		exit 1
	fi

	if [[ $# -lt 1 ]]; then
		targets="check build unit-test integration-test"
	else
		targets=($@)
	fi

	for target in "${targets[@]}"; do
		target "$target"
		ret=$?
		if (( ret != 0 )); then
			return $ret
		fi
	done
	
	if [[ $CAL_INTEGRATION_TEST_COVERAGE == "yes" ]]; then
		if ! echo "${targets[@]}" | grep -q "integration" ; then
			return $ret
		fi
		# 
		# kill pouchd-test and get the coverage
		#
		pkill --signal 3 pouchd-test || echo "no pouchd-test to be killed"
		sleep 5

		tail -1 "$TMP/log"
		cat "$DIR/integrationcover.out" >> "$DIR/coverage.txt"
		return $ret
	fi
}

main "$@"

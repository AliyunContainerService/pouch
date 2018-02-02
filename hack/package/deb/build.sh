#!/usr/bin/env sh

set -e
# This script is to build pouch rpm package as follows,
# Following the below command to build rpm
# 1. Build pouch:deb image
#	cd pouch
#	docker build -t pouch:deb -f hack/package/deb/ubuntu-xenial/Dockerfile.x86_64 .
# 2. Mount a directory which contains gpg keys, eg
#    $ tree /root/deb/
#		deb
#		├── keys
#		│   ├── gpg
#		│   └── secretkey
# 3. run pouch:deb to generate deb package
#	 docker run -it -e VERSION="0.1.0" -v /root/deb/:/root/deb pouch:deb bash -c hack/package/deb/build.sh
# 4. In this example, rpm package will be output in '/root/deb/package/' directory

DIR="$( cd "$( dirname "$0" )" && pwd )"

TMP=$(mktemp -d /tmp/pouch.XXXXXX)

MOUNTDIR=/root/deb
PACKAGEDIR=/root/deb/package
BASEDIR=/go/src/github.com/alibaba/pouch

rm -rf $BASEDIR/debian && cp -r $DIR/debian $BASEDIR/
rm -rf $BASEDIR/systemd && cp -r $DIR/systemd $BASEDIR/

if [ ! -e $$BASEDIR/Makefile.bak ] ;then
	mv $BASEDIR/Makefile $BASEDIR/Makefile.bak
	cp $DIR/Makefile $BASEDIR/Makefile
fi

cd $BASEDIR

debSource="$(awk -F ': ' '$1 == "Source" { print $2; exit }' debian/control)"
debMaintainer="$(awk -F ': ' '$1 == "Maintainer" { print $2; exit }' debian/control)"
debDate="$(date --rfc-2822)"

cat > "debian/changelog" <<-EOF
$debSource (${VERSION}-0~${DISTRO}) $SUITE; urgency=low
  * Version: $VERSION
 -- $debMaintainer  $debDate
EOF

# build package
dpkg-buildpackage -uc -us

# recover Makefile
[ -e $BASEDIR/Makefile.bak ] && mv $BASEDIR/Makefile.bak $BASEDIR/Makefile
mkdir -p $PACKAGEDIR && mv $BASEDIR/../pouch_* $PACKAGEDIR/

# import gpg key
gpg --import $MOUNTDIR/keys/gpg
gpg --import $MOUNTDIR/keys/secretkey

# sign packages
cd $PACKAGEDIR && rm -f Packages.gz Packages
dpkg-scanpackages . /dev/null | gzip -9c > Packages.gz
apt-ftparchive release ./ > Release

# if you want to use your gpg key, you should replace '439AE9EC' with your key
gpg -abs --default-key 439AE9EC -o Release.gpg Release
gpg --clearsign --default-key 439AE9EC -o InRelease Release
echo "packages were output in $PACKAGEDIR "

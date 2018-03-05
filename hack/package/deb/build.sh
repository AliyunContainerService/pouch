#!/usr/bin/env sh

set -e

TMP=$(mktemp -d /tmp/pouch.XXXXXX)
DEFAULT_GPG_KEY=${GPGKEY:-"439AE9EC"}
MOUNTDIR=/root/deb
PACKAGEDIR=/root/deb/package/deb/
BASEDIR=/go/src/github.com/alibaba/pouch

# prepare files that used to build deb packages
rm -rf $BASEDIR/debian && cp -r $BASEDIR/hack/package/deb/debian $BASEDIR/
rm -rf $BASEDIR/systemd && cp -r $BASEDIR/hack/package/deb/systemd $BASEDIR/
rm -f $BASEDIR/Makefile.bak && mv $BASEDIR/Makefile $BASEDIR/Makefile.bak
cp $BASEDIR/hack/package/deb/Makefile $BASEDIR/Makefile

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
gpg -abs --default-key $DEFAULT_GPG_KEY -o Release.gpg Release
gpg --clearsign --default-key $DEFAULT_GPG_KEY -o InRelease Release

echo "Build deb package successfully! Please get the package in $PACKAGEDIR!"
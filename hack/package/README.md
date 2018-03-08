# Building pouch deb and rpm packages 

This file shows how to build Pouch packages.

## Prerequisites

Firstly, you should install Docker, as we use Docker to build images and run containers to build packages.

Then, run `date` command to ensure time is right, which is used to set `BuildTime` field. 

To build Pouch packages, you should prepare gpg keys and put it in a directory.
 
eg:

```bash
$ tree /root/packages/
	packages
	├── keys
	│   ├── gpg
	│   └── secretkey
```

## Building packages

Firstly, checkout to the branch and commit id that used to build packages.

Secondly, run the following command to build Pouch packages, in the progress you will be asked to enter passphrase to sign packages.

What's more, you should pass four parameters as follows:

- `VERSION`, the version of Pouch.

- `ITERATION`, RPM calls this the 'release', FreeBSD calls it 'PORTREVISION'. Debian calls this 'debian_revision'. This parameter is only used to build rpm package.

- `KEYDIR`, the directory of gpg keys.

- `PKGTYPE`, if this parameter is null, we will build rpm and deb packages. You can parse `deb` or `rpm` to build deb or rpm package.

```bash
cd pouch/
./hack/package/package.sh 1.0.0 1.el7.centos /root/test rpm
                            |        |            |      | 
                        VSERSION  ITERATION    KEYDIR  PKGTYPE
```

Finally, packages will be output in `/root/packages/package` in this example.
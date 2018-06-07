# Introduction

In order to provide high quality `pouch`, testing is an important part.

This doc will give an introduction about the following three parts:

* the organization of pouch test
* the usage of pouch test
* the development of pouch test

## Organization of test

Test in pouch could be divided into following parts:

* [unit testing](https://en.wikipedia.org/wiki/Unit_testing#Description)
* [integration testing](https://en.wikipedia.org/wiki/Integration_testing)

*Unit testing* uses [go testing](https://golang.org/pkg/testing/) package, named with `_test.go` suffix and always locates in the same directory with the code tested. [client/client_test.go](https://github.com/alibaba/pouch/blob/master/client/client_test.go) is a good example of unit test.

*Integration test* is in `pouch/test`, programmed with `go language`. There are two kinds of integration test, API test named as `api_xxx_test.go` and command line test named as `cli_xxx_test.go` ("xxx" represents the test point).
It uses [go-check](https://labix.org/gocheck) package, a rich testing framework for go language. It provides many useful functions, such as:

* SetUpTest: Run before each test to do some common work.
* TearDownTest: Run after each test to do some cleanup work.
* SetUpSuite: Run before each suite to do common work for the whole suite.
* TearDownSuite: Run after each suite to do cleanup work for the whole suite.

For other files, they are:

* `main_test.go` : the entrypoint of integration test.
* `utils.go`: common lib functions.
* `environment directory`: directory environment is used to hold environment variables.
* `command package`: package command is used to encapsulate CLI lib functions.
* `request package`: package request is used to encapsulate http request lib functions.

For pouch developer, if your code is only used in a single module, then the unit test is enough. While if your code is called by multiple modules, integration tests are required. In pouch, both of them are developed using go language. More details could be gotten in [Unit Testing](#unit-testing) and [Integration Testing](#integration-testing).

## Run Tests

Tests could be run through the target provided by `Makefile` in the root directory of pouch code. Also could be run manually.
To run the test automatically, the following prerequisites are needed:

    * golang is installed and GOPATH and GOROOT is set correctly
    * docker is installed

Then you could just clone the pouch source to GOPATH and run tests as following:

```
# which docker
/usr/bin/docker
# env |grep GO
GOROOT=/usr/local/go
GOPATH=/go
# cd /go/src/github.com/alibaba/pouch
# make test
```

Using `make -n test`, let us take a look at what `make test` has done.

```
#make -n test
bash -c "env PATH=/sbin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/usr/X11R6/bin:/usr/local/go/bin:/opt/satools:/root/bin hack/make.sh \
check build unit-test integration-test cri-test"
```

`make test` calls the hack/make.sh script to check code format, build pouch daemon and client, run unit test, run integration test and run cri test.
`hack/make.sh` needs `docker` installed on test machine, as it uses `docker build` to build a test image including tools needed to run `make test`. `go` is also needed to be installed and set `GOPATH` `GOROOT` `PATH` correctly. For more information, you could check the `hack/make.sh` script.

## Run Tests Manually

As a pouch developer, you may need to run tests manually.

If you want to do code format check, you could run `make check` in your code directory, but please make sure the following tools are already installed:

* gofmt
* golint
* swagger

In order to run unit-test or integration test, install go and configure go environment first.

```
# go version
go version go1.9.2 linux/amd64
# which go
/usr/local/go/bin//go
# env |grep GO
GOROOT=/usr/local/go
GOPATH=/go
```

Then copy or clone pouch source code to the GOPATH：

```
# pwd
/go/src/github.com/alibaba/pouch
```

Make a build folder to use later:

```
BUILDPATH=/tmp/pouchbuild
export GOPATH=$GOPATH:$BUILDPATH
```

And please notice that files in `/tmp` directory may be deleted after reboot.

Now you could run unit test as following:

```
# make unit-test
```

Or using go test $testdir to run unit test in a specified directory.

```
#go test ./client
ok      github.com/alibaba/pouch/client    0.094s

```

There are more works to do for integration test compared with unit test.

First you need to make sure `pouch` and `pouchd` binary is installed or built.

Then you need to install containerd, runc, lxcfs and so on. You could refer to function `install_pouch` in `hack/make.sh`. There is also a quick way to install them, just install `pouch` package then replace pouch binary with yours.

Next you need to start pouch daemon and pull a busybox image:

```
# pouchd -D --enable-lxcfs=true --lxcfs=/usr/bin/lxcfs >/tmp/log 2>&1 &
# pouch pull registry.hub.docker.com/library/busybox:latest
```

Then integration test could be run as following:

* run entire test:

```
# cd test
#go test
```

* run a single test suite:

```
#go test -check.f PouchCreateSuite
OK: 3 passed
PASS
ok      github.com/alibaba/pouch/test    3.081s
```

* run a single test case:

```
#go test -check.f PouchHelpSuite.TestHelpWorks
OK: 1 passed
PASS
ok      github.com/alibaba/pouch/test    0.488s
```

* run with more information:

```
#go test -check.vv
```

## Development of Test

// TODO

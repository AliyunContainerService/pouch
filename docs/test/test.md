# Introduction

In order to provide high quality `pouch`, testing is an important part.

This doc will give an introduction about the following three parts:

*  the organization of pouch test
*  the usage of pouch test
*  the development of pouch test

# Organization of test

Test in pouch could be divided into following parts:

* [unit testing](https://en.wikipedia.org/wiki/Unit_testing#Description)
* [integration testing](https://en.wikipedia.org/wiki/Integration_testing)

More info could be gotten from the wiki page. 

For pouch developer, if your code is only used in a single module, then the unit test is enough. While if your code is called by multiple modules, integration tests are required. In pouch, both of them are developed using go language. More details could be gotten in [Unit Testing](#unit-testing) and [Integration Testing](#integration-testing).

# Usage of Test

Tests could be run through the target provided by `Makefile` in the root directory of pouch code. Also could be run manually.	
To run the whole test, you could just run:

```
# make test
```
Please note that, in order to run test, the following prerequisites are needed:
	
	* golang is installed
	* docker is installed


## Unit Testing

Unit testing uses [go testing](https://golang.org/pkg/testing/) package, named with `_test.go` postfix and always locates in the same directory with the code tested. [client/client_test.go](https://github.com/alibaba/pouch/blob/master/client/client_test.go) is a good example of unit test.

There are two ways to trigger unit test. 

* Using [Makefile](https://github.com/alibaba/pouch/blob/master/Makefile) target unit-test to run entire unit test.

	```
	# make unit-test
	``` 

* Using go test $testdir to run unit test in a specified directory.

	```
	#go test ./client
	ok  	github.com/alibaba/pouch/client	0.094s

	```
Note: If you are testing modules importing "github.com/docker/libnetwork/xxx", there may be errors like following:

	```
	#go test ./daemon/mgr/
	../../docker/libnetwork/config/config.go:6:2: cannot find package "github.com/BurntSushi/toml" in any of:
		/usr/local/go/src/github.com/BurntSushi/toml (from $GOROOT)
		/home/sit/letty/src/github.com/BurntSushi/toml (from $GOPATH)
		<snip>
	```
In this case, you may have to run unit test from Makefile. In this way, it will call `hack/build` script, and do setup related to libnetwork package location before running unit test. 

## Integration Testing

Integration test is in `pouch/test`, programmed with `go language`. There are two kinds of integration test, API test named as `api_xxx_test.go` and command line test named as `cli_xxx_test.go` ("xxx" represents the test point). 

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


There are two ways to trigger integration test. 

* Using [Makefile](https://github.com/alibaba/pouch/blob/master/Makefile) target integration-test to run entire integration test.

	```
	# make integration-test
	```
* Using go test to run integration test. Before running integration test, users need to [build pouch](https://github.com/alibaba/pouch/blob/master/INSTALLATION.md) binary and launch `pouchd` daemon.
Then integration test could be run as following:

	* run entire test: 

		```
		#go test ./test
		```
	* run a single test suite: 
	
		```
		#go test -check.f PouchCreateSuite
		OK: 3 passed
		PASS
		ok  	github.com/alibaba/pouch/test	3.081s
		```
	* run a single test case:
	
		```
		#go test -check.f PouchHelpSuite.TestHelpWorks
		OK: 1 passed
		PASS
		ok  	github.com/alibaba/pouch/test	0.488s	
		```	
	* run with more information:
	
		```
		#go test -check.vv 
		```
	
# Development of Test
// TODO

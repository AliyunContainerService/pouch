# Code Style

Code style is a set of rules or guidelines when writing source codes of a software project. Following particular code style will definitely help contributors to read and understand source codes. In addition, it will help to avoid introducing errors as well.

## Code Style Tools

Project Pouch is written in Golang. And currently we use three tools to help conform code styles of this project. These three tools are:

* [gofmt](https://golang.org/cmd/gofmt)
* [golint](https://github.com/golang/lint)
* [go vet](https://golang.org/cmd/vet/)

And all these tools are used in [Makefile](../../Makefile).

## Code Review Comments

When collaborating in Pouch project, we follow the style from [Go Code Review Commnets](https://github.com/golang/go/wiki/CodeReviewComments). Before contributing, we treat this as a must-read.

## Additional Style Rules

For a project, existing tools and rules may not be sufficient. To align more in styles, we recommend contributors taking a thorough look at the following additional style rules:

* When constructing a struct, if comments needed for fields in struct, keep a blank line between fields;
* When defining interface functions, we should always explicitly add formal parameter, and this helps a lot to code readability;
* When defining interface functions, if comments needed for functions, keep a blank line between functions;
* When importing packages, to improve readabilities, we should import package by sequence: system packages, project's own packages and third-party packages. And we should keep a blank line among these three kinds of packages;
* Const object should be declared at the beginning of the go file, following package name and importing;
* When generating error in an action failure, we should generally use the way of `fmt.Errorf("failed to do something: %v", err)`;
* No matter log or error, first letter of the message must be lower-case;
* When occuring nesting errors, we recommend first considering using package `github.com/pkg/errors`;
* Every comment should begin with `//` plus a space, and please don't forget the whitespace, and end with a `.`;
* We should take `DRY(Don't Repeat Yourself)` into more consideration.

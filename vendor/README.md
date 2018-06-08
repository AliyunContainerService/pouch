# Vendor

## Tools

Pouch uses tool [govendor](https://github.com/kardianos/govendor) to vendor packages in its own repository.

## Installation

You could use the common way to install `govendor`:

```bash
go get -u github.com/kardianos/govendor
```

Pouch chooses version v1.0.8 to be the collaborate. 

## Workflow

### Initialization

It is really easy to utilize `govendor` to initialize vendor policy in a project:

```bash
cd pouch
govendor init
```

### Vendor External Packages

In pouch, when vendoring a package, we take the version of it very seriously. Therefore, almostly every time we use `govendor`, we explicitly pick the specified commit ID or tag.

As a result, we seldom use the command `govendor add github.com/a/b`. Instead, we always make use of `govendor add github.com/a/b@8ac97434144a8b58bd39576fc65be42d90257b04` or `govendor add github.com/a/b@v1.3.0`.

When using `govendor`, there are still some tiny tips for us:

* If you pick commit id `8ac97434144a8b58bd39576fc65be42d90257b04` of package `govendor add github.com/a/b`, then for this package in your `GOPATH`, you have to checkout to this specified commit ID before using `govendor`.
* When there are go files and directories which are all you need to vendor under the root path of package `govendor add github.com/a/b`, you have to use flag `^` to include them all using command `govendor add github.com/a/b/^@8ac97434144a8b58bd39576fc65be42d90257b04`.
* Assuming that package `github.com/a/b` also has a directory `vendor`, so-called nesting vendor, in most cases we remain these packages.

### Remove Vendored Packages

Removal of vendored package will take no effort. Just execute the following the command:

```bash
govendor remove github.com/a/b
```

### Update Vendored Packages

Update of vendored packages will take a little bit more effort than removal. Although you could execute `govendor update github.com/a/b@a4bbce9fcae005b22ae5443f6af064d80a6f5a55` in which `a4bbce9fcae005b22ae5443f6af064d80a6f5a55` is the new commit ID, you should still keep it in mind that for this package in your `GOPATH`, you have to checkout to this specified commit ID before using `govendor update`.

## Resources

More details, you could refer to [govendor](https://github.com/kardianos/govendor).




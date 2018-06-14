# PouchContainer Command Line

You can take advantage of PouchContainer command line to experience the power of PouchContainer. If you are not familiar with PouchContainer yet, command line document is a must-read for you at the very beginning. Of course, you make sure that you have installed PouchContainer. More installation details, you can refer to [INSTALLATION.md](../../INSTALLATION.md).

PouchContainer has an architecture of client/server, thus PouchContainer has two part of command line tools:

* pouchd: a daemon side binary to run as a server;
* pouch: a client side binary to run as a client.

## pouchd

You can set up PouchContainer daemon by executing `pouchd` which is located in your `$PATH` if you have installed PouchContainer correctly. The running pouchd process can accept requests from PouchContainer CLI, handle requests and manage containers. `pouchd` is a long-running process background, and you can config for it by passing command line flags which is defined in pouchd.

## PouchContainer

You can use client side tool `pouch` to interact with daemon side process `pouchd`. Flags and arguments can be input to do what actually you wish. Then PouchContainer parses the flags and arguments and sends a RESTful request to daemon side `pouchd`.

If we divide the functionality into parts, we can conclude the following parts:

* Images Management
* Container Management
* Storage Management
* Network Management
* System Management

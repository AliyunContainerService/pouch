# Pouch Command Line

You can take advantage of pouch command line to experience the power of pouch. If you are not familiar with pouch yet, command line document is a must-read for you at the very beginning. Of cource, you make sure that you have installed pouch. More installation details, you can refer to [INSTALLATION.md](../../INSTALLATION.md).

Pouch has an architecture of client/server, thus pouch has two part of command line tools:

* pouchd: a daemon side binary to run as a server;
* pouch: a client side binary to run as a client.

## pouchd

You can set up pouch daemon by executing `pouchd` which is located in your `$PATH` if you have installed pouch correctly. The running pouchd process can accept requests from pouch cli, handle requests and manage containers. `pouchd` is a long-running process background, and you can config for it by passing command line flags which is defined in pouchd.

## pouch

You can use client side tool `pouch` to interact with daemon side process `pouchd`. Flags and arguments can be input to do what actually you wish. Then pouch parses the flags and arguments and sends a RESTful request to daemon side `pouchd`.

If we divide the fuctionality into parts, we can conclude the following parts:

* Images Management
* Container Management
* Storage Management
* Network Management
* System Management

# Vagrant Setup to Test the PouchContainer

This documentation highlights how to use [Vagrant](https://www.vagrantup.com/) to setup node to test PouchContainer.

## Pre-requisites

This was tested on:

- Vagrant 1.9.5
- VirtualBox 5.1.26

## Getting Started

Clone this repo, change to the `hack/vagrant` directory and let Vagrant do the work.

    $ vagrant up
    $ vagrant status
    Current machine states:

    pouch-dev-node            running (virtualbox)

You are now ready to SSH to `pouch-dev-node`, source code of PouchContainer is mounted to /go/src/github.com/alibaba/pouch on `pouch-dev-node`.

    $ vagrant ssh pouch-dev-node
    vagrant@pouch-dev-node:~$ sudo su
    root@pouch-dev-node:/home/vagrant# cd /go/src/github.com/alibaba/pouch/
    root@pouch-dev-node:/go/src/github.com/alibaba/pouch# make && make test
    ...<snip>...


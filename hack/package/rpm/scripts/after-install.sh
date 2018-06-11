#!/usr/bin/env bash
if [ "$1" -eq 1 ] ; then
	systemctl preset pouch > /dev/null 2>&1
	
fi
if ! getent group pouch > /dev/null; then
	groupadd --system pouch
fi

if [ ! -d "/var/lib/pouch-lxcfs" ] ; then
    mkdir -p /var/lib/pouch-lxcfs
fi
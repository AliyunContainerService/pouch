#!/usr/bin/env bash
if [ "$1" -gt 0 ] ; then
	rm -f /var/lib/rpm-state/pouch-is-active > /dev/null 2>&1
	if systemctl is-active pouch > /dev/null 2>&1 ; then
		systemctl stop pouch > /dev/null 2>&1
		touch /var/lib/rpm-state/pouch-is-active > /dev/null 2>&1
	fi
fi

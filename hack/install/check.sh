#!/usr/bin/env bash

# detect_os return linux OS distribution.
#
# NOTE: only support Ubuntu/Debian/Centos/Redhat
detect_os() {
  local os dist

  os="$(uname -s | tr "[:upper:]" "[:lower:]")"

  if [[ "${os}" = "linux" ]]; then
    if [ -f /etc/lsb-release -o -d /etc/lsb-release.d ]; then
      dist="$(lsb_release -i | cut -d: -f2 | sed s/'^\t'//)"
    elif [ -f /etc/redhat-release -o -f /etc/centos-release ]; then
      dist="RedHat"
    fi
  fi

  if [ -z "${dist}" ]; then
    echo >&2 "failed to detect os distribution"
    exit 1
  fi
  echo "${dist}"
}

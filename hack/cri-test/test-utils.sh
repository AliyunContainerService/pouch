#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
. ${ROOT}/versions
# POUCH_FLAGS are the extra flags to use when start pouchd.
POUCH_FLAGS=${POUCH_FLAGS:-""}
# RESTART_WAIT_PERIOD is the period to wait before restarting pouchd/containerd.
RESTART_WAIT_PERIOD=${RESTART_WAIT_PERIOD:-10}

POUCH_SOCK=/var/run/pouchcri.sock

pouch_pid=
containerd_pid=

# test_setup starts containerd and pouchd.
test_setup() {
  local report_dir=$1 

  # Start containerd
  if [ ! -x "$(command -v containerd)" ]; then
    echo "containerd is not installed, please run hack/make.sh"
    exit 1
  fi
  containerd_pid_command=`pgrep containerd`
  containerd_pid=${containerd_pid_command}
  if [ ! -n "${containerd_pid}" ]; then
    keepalive "/usr/local/bin/containerd" ${RESTART_WAIT_PERIOD} &> ${report_dir}/containerd.log &
    containerd_pid=$!
  fi
  # Wait for containerd to be running by using the containerd client ctr to check the version
  # of the containerd server. Wait an increasing amount of time after each of five attempts
  readiness_check "ctr version"

  # Start pouchd
  pouch_pid_command=`pgrep pouchd`
  pouch_pid=${pouch_pid_command}
  if [ ! -n "${pouch_pid}" ]; then
    keepalive "pouchd ${POUCH_FLAGS}" \
	  ${RESTART_WAIT_PERIOD} &> ${report_dir}/pouch.log &
    pouch_pid=$!
  fi
  readiness_check "pouch version"
}

# test_teardown kills containerd and cri-containerd.
test_teardown() {
  if [ -n "${containerd_pid}" ]; then
    kill ${containerd_pid}
  fi
  if [ -n "${pouch_pid}" ]; then
    kill ${pouch_pid}
  fi
  sudo pkill containerd
}

# keepalive runs a command and keeps it alive.
# keepalive process is eventually killed in test_teardown.
keepalive() {
  local command=$1
  echo ${command}
  local wait_period=$2
  while true; do
    ${command}
    sleep ${wait_period}
  done
}

# readiness_check checks readiness of a daemon with specified command.
readiness_check() {
  local command=$1
  local MAX_ATTEMPTS=5
  local attempt_num=1
  until ${command} &> /dev/null || (( attempt_num == MAX_ATTEMPTS ))
  do
      echo "$attempt_num attempt \"$command\"! Trying again in $attempt_num seconds..."
      sleep $(( attempt_num++ ))
  done
}


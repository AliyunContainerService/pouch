#!/usr/bin/env bash

# integration::run_local_persist_background runs local-persist in background.
integration::run_local_persist_background() {
  local log_file
  log_file=$1

  echo "start local-persist volume plugin..."
  local-persist > "${log_file}" 2 >&1 &
}

# integration::stop_local_persist stop local-persist.
integration::stop_local_persist() {
  echo "stop local-persist volume plugin..."
  set +e
  pkill local-persist
  rm -rf /run/docker/plugins/local-persist.sock
  set -e
}

# integration::run_mount_lxcfs_background runs lxcfs mount to /var/lib/lxcfs.
integration::run_mount_lxcfs_background() {
  echo "start mount lxcfs /var/lib/lxcfs..."
  lxcfs /var/lib/lxcfs 2>&1 &
}

# integration::stop_mount_lxcfs stop lxcfs mount.
integration::stop_mount_lxcfs() {
  echo "stop lxcfs /var/lib/lxcfs..."
  set +e; pkill lxcfs; rm -rf /run/lxcfs.pid; set -e
}

# integration::run_pouchd_background runs pouchd in background.
integration::run_pouchd_background() {
  echo "start pouch daemon..."

  local cmd flags log_file

  cmd="$1"
  flags="$2"
  log_file="$3"

  cmd="${cmd} ${flags}"
  ${cmd} > "${log_file}" 2>&1 &
}

# integration::stop_pouchd stops pouchd.
integration::stop_pouchd() {
  echo "stop pouch daemon..."
  set +e; pkill -3 pouchd; set -e
}

# integration::ping_pouchd makes sure that pouchd started.
integration::ping_pouchd() {
  local timeout

  # make sure that it's up.
  timeout=30
  while true;
  do
    if pouch version 2> /dev/null; then
      break
    elif (( $((timeout--)) == 0 ));then
      echo "failed to start pouch daemon in background!"
      return 1
    fi
    sleep 1
  done
}

#!/bin/bash

set -ex

# update the library cache
ldconfig

echo "[entrypoint.sh] launching zfs-driver."

/usr/local/bin/zfs-driver "$@" &

child=$!

#sigterm caught SIGTERM signal and forward it to child process
_sigterm() {
  echo "[entrypoint.sh] caught SIGTERM signal forwarding to pid [$child]."
  kill -TERM "$child" 2> /dev/null
  waitForChildProcessToFinish
}

#sigint caught SIGINT signal and forward it to child process
_sigint() {
  echo "[entrypoint.sh] caught SIGINT signal forwarding to pid [$child]."
  kill -INT "$child" 2> /dev/null
  waitForChildProcessToFinish
}

#waitForChildProcessToFinish waits for child process to finish
waitForChildProcessToFinish(){
    while ps -p "$child" > /dev/null; do sleep 1; done;
}

trap _sigint INT
trap _sigterm SIGTERM

wait $child

#!/bin/bash
set -ex

COMMAND="${@:-start}"

function start () {
  exec ironic-api \
        --config-file /etc/ironic/ironic.conf
}

function stop () {
  kill -TERM 1
}

$COMMAND


#!/bin/bash
set -ex

ironic-dbsync --debug --config-file=/etc/ironic/ironic.conf upgrade


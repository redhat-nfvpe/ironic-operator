#!/bin/bash
set -ex

mkdir -p /var/lib/pod_data/ironic/images
mkdir -p /var/lib/pod_data/master_images
mkdir -p /var/lib/pod_data/ironic_images

exec ironic-conductor \
      --config-file /etc/ironic/ironic.conf \
      --config-file /tmp/pod-shared/conductor-local-ip.conf

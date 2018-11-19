#!/bin/bash
set -ex

ln -s /var/lib/pod_data/tftpboot /tftpboot

chown -R nobody /tftpboot
chmod -R a+rx /tftpboot

exec /usr/sbin/in.tftpd \
  --verbose \
  --verbosity 7 \
  --foreground \
  --user root \
  --address 0.0.0.0:69 \
  --map-file /tftp-map-file /tftpboot

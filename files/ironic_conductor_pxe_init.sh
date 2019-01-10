#!/bin/bash
set -ex

mkdir -p /var/lib/pod_data/tftpboot
mkdir -p /var/lib/pod_data/tftpboot/master_images

for FILE in undionly.kpxe ipxe.efi; do
  if [ -f /usr/share/ipxe/$FILE ]; then
    cp -v /usr/share/ipxe/$FILE /var/lib/pod_data/tftpboot
  fi
done

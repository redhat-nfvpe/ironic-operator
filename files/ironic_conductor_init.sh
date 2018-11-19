#!/bin/bash
set -ex

if [ "x" == "x${PXE_NIC}" ]; then
    # just use the first available nic
    $PXE_NIC=$(ip link | awk -F: '$0 !~ "lo|vir|wl|^[^0-9]"{print $2;getline}' | head -n 1 | sed -e 's/^[[:space:]]*//')
fi

function net_pxe_ip {
    net_pxe_addr=$(ip addr | awk "/inet / && /${PXE_NIC}/{print \$2; exit }")
    echo $net_pxe_addr | awk -F '/' '{ print $1; exit }'
}
PXE_IP=$(net_pxe_ip)

if [ "x" == "x${PXE_IP}" ]; then
    echo "Could not find IP for pxe to bind to"
    exit 1
fi

tee /tmp/pod-shared/conductor-local-ip.conf << EOF
[DEFAULT]

# IP address of this host. If unset, will determine the IP
# programmatically. If unable to do so, will use "127.0.0.1".
# (string value)
my_ip = ${PXE_IP}

[pxe]
# IP address of ironic-conductor node's TFTP server. (string
# value)
tftp_server = ${PXE_IP}

[deploy]
# ironic-conductor node's HTTP server URL. Example:
# http://192.1.2.3:8080 (string value)
# from .deploy.ironic.http_url
http_url = http://${PXE_IP}:8081
EOF

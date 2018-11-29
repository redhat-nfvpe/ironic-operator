#! /bin/bash
yum install -y dhcp

dhcpd -f -d -cf /data/dhcpd.conf -user dhcpd -group dhcpd --no-pid ${PXE_NIC}


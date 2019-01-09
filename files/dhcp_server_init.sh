#!/bin/bash
set -ex

# strip hosts in lines and output content
if [[ ! -z $DHCP_HOSTS ]]; then
    echo "deny unknown-clients;" >> /data/hosts/hosts.conf

    COUNTER=1
    while read -r line; do
      echo "host host${COUNTER} { hardware ethernet ${line}; }" >> /data/hosts/hosts.conf
      COUNTER=$[$COUNTER+1]
    done <<< "${DHCP_HOSTS}"
else
    touch /data/hosts/hosts.conf
fi

# now generate the entry for the zone
if [ "x" == "x${PXE_NIC}" ]; then
    # just use the first available nic
    $PXE_NIC=$(ip link | awk -F: '$0 !~ "lo|vir|wl|^[^0-9]"{print $2;getline}' | head -n 1 | sed -e 's/^[[:space:]]*//')
fi

function net_pxe_ip {
    net_pxe_addr=$(ip addr | awk "/inet / && /${PXE_NIC}/{print \$2; exit }")
    echo $net_pxe_addr | awk -F '/' '{ print $1; exit }'
}
PXE_IP=$(net_pxe_ip)

# given the ip, extract subnet
IP_FRAGMENT=${PXE_IP%.*}
SUBNET="${IP_FRAGMENT}.0"
BROADCAST="${IP_FRAGMENT}.255"

tee /data/zones/zone.conf << EOF
subnet ${SUBNET} netmask 255.255.255.0 {
  option subnet-mask 255.255.255.0;
  option routers ${PXE_IP};
  option broadcast-address ${BROADCAST};
  option domain-name-servers ${PXE_IP};
  option domain-name "${CLUSTER_DOMAIN}";
  range dynamic-bootp ${IP_FRAGMENT}.${INITIAL_IP_RANGE} ${IP_FRAGMENT}.${FINAL_IP_RANGE};
  default-lease-time 21600;
  max-lease-time 43200;

  # ipxe boot
  if exists user-class and option user-class = "iPXE" {
      filename "http://${PXE_IP}:8081/boot.ipxe";
  } else {
      if option client-arch != 00:00 {
          filename "ipxe.efi";
      } else {
          filename "undionly.kpxe";
      }

      next-server ${PXE_IP};
  }
}

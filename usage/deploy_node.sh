#!/bin/bash

# to be executed inside operator pod
export OS_TOKEN=fake-token
export OS_URL=http://openstack-ironic:6385

openstack baremetal node list

# TODO: retrieve right ipmi creds
export NODE_IPMI_ADDRESS="192.168.126.1"
export NODE_IPMI_USERNAME="admin"
export NODE_IPMI_PASSWORD="password"
export NODE_IPMI_PORT=6232
export NODE_PXE_MAC="pxe:mac:address"

export IMAGES_SERVER="ip-from-server"
export DEPLOY_KERNEL="http://$IMAGES_SERVER/ironic-python-agent.kernel"
export DEPLOY_RAMDISK="http://$IMAGES_SERVER/ironic-python-agent.initramfs"

openstack baremetal node create --driver ipmi --driver-info ipmi_address=$NODE_IPMI_ADDRESS \
    --driver-info ipmi_username=$NODE_IPMI_USERNAME \
    --driver-info ipmi_password=$NODE_IPMI_PASSWORD \
    --driver-info ipmi_port=$NODE_IPMI_PORT  \
    --driver-info deploy_kernel=$DEPLOY_KERNEL \
    --driver-info deploy_ramdisk=$DEPLOY_RAMDISK

# TODO: properly retrieve generated node uuid
export NODE_UUID=dummy_uuid
openstack baremetal port create $NODE_PXE_MAC --node $NODE_UUID
openstack baremetal node validate $NODE_UUID
openstack baremetal node manage $NODE_UUID
openstack baremetal node provide $NODE_UUID

# now generate config drive
mkdir -p /tmp/config-drive/openstack/latest
curl http://$IMAGES_SERVER/artifacts/stable_ignition/dummy.ign -o user_data
yum install -y genisoimage # we may need this dep

# and now deploy
openstack baremetal node deploy $NODE_UUID --config-drive /tmp/config-drive/

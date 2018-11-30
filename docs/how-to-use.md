# How to use

This Ironic operator is providing you with a simple [standalone Ironic] that you can use for simple baremetal deployments. The available deploy modes are [direct] and [ramdisk].
There is no authentication system like Keystone enabled, so you will need to take care of protecting your endpoints on your own. Also no network services like Neutron are enabled, the PXE boot is achieved by this operator, providing an ISC DHCP server that can be configured.

## How to communicate with the API
First thing that you need to do is download latest [python-openstackclient] and [python-ironicclient] . Once those clients are installed, you will be able to interact with `openstack baremetal` commands.
Second thing is to setup the authentication. As we haven't installed Keystone, it simply needs a fake token to be exported:
```sh
export OS_TOKEN=fake-token
```
The Ironic endpoint is set to the name `openstack-ironicapi`. If you ran the API command from a server inside the cluster, your DNS will have been automatically configured. Otherwise, you need to check for the matching IP of the Ironic API service, and configure your DNS properly. Then, export the endpoint:
```sh
export OS_URL=http://openstack-ironicapi:6385
```
Once authentication is done, you can proceed with direct or ramdisk deployment of baremetals.

## Procedure for Ironic direct deploy
- To deploy a new server, first thing needed is to create the node. To create a new node you will need to know IPMI address (and port if needed), and also the credentials to access (ipmi_user and ipmi_password).
- You will also need to have the [Ironic Python Agent] image. If you want to use the RDO images you can download them from [RDO Ironic Python Agent image]. You will need to store those images in an HTTP server, and your Ironic pods need to have access to it.
- When you have the IPA and the IPMI credentials, is time to create the node. It can be done with:
```sh
openstack baremetal node create --driver ipmi --driver-info ipmi_address=<ipmi_address> --driver-info ipmi_username=<ipmi_user> --driver-info ipmi_password=<ipmi_password> [--driver-info ipmi_port=<ipmi_port>]  --driver-info deploy_kernel=[http|https]://<http_server>/ironic-python-agent.kernel --driver-info deploy_ramdisk=[http|https]://<http_server>/ironic-python-agent.initramfs
```
- You can check the results of the node creation with:
```sh
openstack baremetal node list
openstack baremetal node show <node_uuid>
```
- As we don't have neutron, the network port needs to be created manually. To achieve that, you will need to have the MAC address of the sever you want to be deployed. This MAC address needs to match with the NIC you are going to use for PXE. You can check that on your server management console, or checking the BIOS of the server. Once you know that MAC, proceed to creating the port with:
```sh
openstack baremetal port create <pxe_mac_address> --node <node_uuid>
``
- Next step, is to provide the information for the deployment image. To achieve that, you will need to have the deployment image (a qcow2 file), stored in same http server you used to store IPAs. You need also to specify the size of the disk that you want to use (you can check your disk size and make it match). Also, for security and validation, you need to provide the md5 checksum of the deployment image. Once you have this information, set the properties:
```sh
openstack baremetal node set <node_uuid> --instance-info image_source=[http|https]://<http_server>/deployment_image.qcow2 --instance-info root_gb=<size_of_disk_in_gb> --instance-info  image_checksum=<deployment_image_checksum>
```
- Then you are ready to validate the node, if that passes you are ready to deploy:
```sh
openstack baremetal node validate <node_uuid>
```
- Once the node has been validated, you need to make it pass through the manage and provide steps. As it is on `standalone` mode, this needs to be done manually:
```sh
openstack baremetal node manage <node_uuid>
openstack baremetal node provide <node_uuid>
```
- And then, finally deploy it:
```sh
openstack baremetal node deploy <node_uuid>
```
- In case deploy failed and you want to redeploy, you can achieve it with this cycle:
```sh
openstack baremetal node undeploy <node_uuid>
openstack baremetal node set <node_uuid> --instance-info image_source=[http|https]://<http_server>/deployment_image.qcow2 --instance-info root_gb=<size_of_disk_in_gb> --instance-info  image_checksum=<deployment_image_checksum>
openstack baremetal node deploy <node_uuid>
```

[standalone Ironic]: <https://docs.openstack.org/ironic/latest/install/standalone.html>
[direct]: <https://docs.openstack.org/ironic/latest/admin/interfaces/deploy.html#direct-deploy>
[ramdisk]: <https://docs.openstack.org/ironic/latest/admin/interfaces/deploy.html#ramdisk-deploy>
[python-openstackclient]: <https://docs.openstack.org/python-openstackclient/latest/>
[python-ironicclient]: <https://docs.openstack.org/python-ironicclient/latest/>
[Ironic Python Agent]: <https://docs.openstack.org/ironic-python-agent/latest/>
[RDO Ironic Python Agent image]: <https://images.rdoproject.org/queens/delorean/current-tripleo/>

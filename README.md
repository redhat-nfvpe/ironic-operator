# Ironic Operator

This project is integrating an standalone [Ironic] deployment natively into a Kubernetes cluster, using the new [Kubernetes operators]
This is a very opinionated install of [Ironic], that will just install the needed components without any other OpenStack dependencies.
At the moment, this operator installs Ironic conductor and API.

# Pre-requisites
In order to successfully deploy Ironic, you will need several requirements:
 - A working MySQL/MariaDB cluster, with root access, to create new user accounts
 - A working RabbitMQ cluster, with permissions to create new accounts and vhosts. It needs to have rabbitmq-management plugin enabled
 - In case you want to persist MySQL and RabbitMQ data, you will need to use some shared storage, like NFS. Please configure those services according to your own needs. Examples of how to deploy those services will be provided
 - Label the node(s) where you are going to deploy Ironic with "ironic-control-plane=enabled"
 - Special permissions on your kubernetes cluster. The account that you are using to run [Ironic] operator on it, needs to have following permissions:

```sh
    - '{"allowHostDirVolumePlugin": true}'
    - '{"allowHostNetwork": true}'
    - '{"allowHostIPC": true}'
    - '{"allowHostPorts": true}'
    - '{"allowPrivilegedContainer": true}'
    - '{"allowedCapabilities": ["DAC_READ_SEARCH", "SYS_RESOURCE", "CAP_NET_RAW", "CAP_NET_ADMIN"]}'
    - '{"requiredDropCapabilities": ["KILL", "MKNOD"]}'
    - '{"runAsUser": {"type": "RunAsAny" }}'
    - '{"seLinuxContext": {"type": "RunAsAny"}}'
    - '{"volumes": ["configMap", "downwardAPI", "emptyDir", "hostPath", "persistentVolumeClaim", "projected", "secret", "nfs"]}'
 ```
  - TEMPORARY: If you are deploying on systems with selinux enabled, please set it to permissive. This won't be needed in the future

# Installation
Once you are in a cluster, clone the [Ironic] operator from [https://github.com/redhat-nfvpe/ironic-operator] . After that, you need to use the manifests on the deploy folder:

```sh
kubectl apply -f crds/ironic_v1alpha1_ironicapi_crd.yaml
kubectl apply -f crds/ironic_v1alpha1_ironicconductor_crd.yaml
kubectl apply -f service_account.yaml
kubectl apply -f role.yaml
kubectl apply -f role_binding.yaml
kubectl apply -f operator.yaml
kubectl apply -f credentials.yaml
kubectl apply -f settings.yaml
kubectl apply -f service_account.yaml
kubectl apply -f crds/ironic_v1alpha1_ironicapi_cr.yaml
kubectl apply -f crds/ironic_v1alpha1_ironicconductor_cr.yaml
```

# Configuration
The operator comes with two different files that are going to be used to configure the Ironic operator: `credentials.yaml` and `settings.yaml`.

* `credentials.yaml`: It contains different secrets that are used to configure the MySQL and Rabbit connection. The secrets are the following:
  - ironic-db-user: this will store the credentials to be used for the ironic account. The [Ironic] operator will create a new account on your MySQL cluster according to it. You need to pass DB_HOST, DB_USER, DB_DATABASE and DB_PASSWORD into Opaque format.
  - mysql-root-credentials: these are the settings to connect to the existing MySQL cluster with root permissions. These settings will be used to create the new Ironic account, that will be the one used by the operator. You need to pass ROOT_DB_HOST, ROOT_DB_USER and ROOT_DB_PASSWORD in Opaque format.
  - ironic-rabbitmq-admin: this is the connection to access to the existing RabbitMQ cluster with root access. It will be used to create an ironic account, that will be used by the operator. You need to pass the whole RABBITMQ_CONNECTION in Opaque format.
  - ironic-rabbitmq-user: this is the connection string used for ironic to access the Rabbit cluster. This will be created by the operator. You need to pass the whole RABBITMQ_CONNECTION in Opaque format. You will also need to pass the RABBITMQ_TRANSPORT, that needs to be the same as RABBITMQ_CONNECTION but pointing to the TCP port.

The secrets are stored in `Opaque` format , you can achieve it by just executing
```sh
echo 'content-of-the-secret' | base64
```

```sh
---
apiVersion: v1
kind: Secret
metadata:
  name: ironic-db-user
type: Opaque
data:
  DB_HOST: bXlzcWw=           # mysql
  DB_USER: aXJvbmlj           # ironic
  DB_DATABASE: aXJvbmlj       # ironic
  DB_PASSWORD: cGFzc3dvcmQ=   # password
---
apiVersion: v1
kind: Secret
metadata:
  name: "mysql-root-credentials"
type: Opaque
data:
  ROOT_DB_HOST: bXlzcWw=           # mysql
  ROOT_DB_USER: cm9vdA==           # root
  ROOT_DB_PASSWORD: cGFzc3dvcmQ=   # password
---
apiVersion: v1
kind: Secret
metadata:
  name: ironic-rabbitmq-admin
type: Opaque
data:
  RABBITMQ_CONNECTION: cmFiYml0Oi8vZ3Vlc3Q6Z3Vlc3RAcmFiYml0bXE6MTU2NzIvaXJvbmlj  # rabbit://guest:guest@rabbitmq:15672/ironic
---
apiVersion: v1
kind: Secret
metadata:
  name: ironic-rabbitmq-user
type: Opaque
data:
  RABBITMQ_CONNECTION: cmFiYml0Oi8vaXJvbmljOnBhc3N3b3JkQHJhYmJpdG1xOjE1NjcyL2lyb25pYw==  # rabbit://ironic:password@rabbitmq:15672/ironic
  RABBITMQ_TRANSPORT: cmFiYml0Oi8vaXJvbmljOnBhc3N3b3JkQHJhYmJpdG1xOjU2NzIvaXJvbmljCg==  # rabbit://ironic:password@rabbitmq:5672/ironic
```
* `settings.yaml`: It contains two different ConfigMaps with settings that can be adjusted for your environment. First one, `pxe-settings` is used to define the NIC settings for PXE boot: You will need to provide:
  - PXE_NIC: eth0 . This will be the nic used for PXE booting in the DHCP server created for Ironic

  Second one, dhcp-settings, will contain the settings to control the DHCP service. You will need to provide:
  - USE_EXTERNAL_DHCP: boolean. Set to True if you are going to use an external DHCP server, so it will skip the deployment of the internal one.
  - CLUSTER_DOMAIN: it needs to match the domain for your Kubernetes cluster, in order for DNS to work
  - INITIAL_IP_RANGE, FINAL_IP_RANGE: It will take network CIDR from the PXE_NIC definition, but this will limit the range of IPS to be assigned for DHCP in PXE boot.
  - DHCP_HOSTS: list that need to match all the MACs for the server that you want to provision with this ironic operator. It is used to don't add PXE boot to all the servers in the system, but just to the ones that we are interested on. If you don't want to limit by hosts, just set `DHCP_HOSTS: ""`

```sh
---
apiVersion: v1
kind: ConfigMap
metadata:
    name: pxe-settings
data:
    PXE_NIC: eth0
---
apiVersion: v1
kind: ConfigMap
metadata:
    name: dhcp-settings
data:
    USE_EXTERNAL_DHCP: False
    CLUSTER_DOMAIN: "edge.testing"
    INITIAL_IP_RANGE: "20"
    FINAL_IP_RANGE: "200"
    DHCP_HOSTS:
        52:54:00:CC:CC:03

```

# How to use

Please follow the [How to use documentation]

[Ironic]: <https://wiki.openstack.org/wiki/Ironic>
[Kubernetes operators]: <https://github.com/operator-framework/operator-sdk>
[https://github.com/redhat-nfvpe/ironic-operator]: <https://github.com/redhat-nfvpe/ironic-operator>
[How to use documentation]: <./docs/how-to-use.md>

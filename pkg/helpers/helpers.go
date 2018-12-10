// Copyright 2018 Red Hat Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helpers

import (
    "context"
    "fmt"
    "log"
    "strings"
    packr "github.com/gobuffalo/packr/v2"

    appsv1 "k8s.io/api/apps/v1"
    batchv1 "k8s.io/api/batch/v1"
    v1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/types"
    "k8s.io/apimachinery/pkg/util/intstr"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

func GetIronicBinConfigMap(namespace string) (*v1.ConfigMap, error) {
    // read all bin scripts
    box := packr.New("files", "../../files")

    db_init, err := box.FindString("db_init.py")
    if err != nil {
        log.Fatal(err)
    }
    db_sync, err := box.FindString("db_sync.sh")
    if err != nil {
        log.Fatal(err)
    }
    rabbit_init, err := box.FindString("rabbit_init.sh")
    if err != nil {
        log.Fatal(err)
    }
    ironic_api, err := box.FindString("ironic_api.sh")
    if err != nil {
        log.Fatal(err)
    }
    ironic_conductor, err := box.FindString("ironic_conductor.sh")
    if err != nil {
        log.Fatal(err)
    }
    ironic_conductor_init, err := box.FindString("ironic_conductor_init.sh")
    if err != nil {
        log.Fatal(err)
    }
    ironic_conductor_pxe, err := box.FindString("ironic_conductor_pxe.sh")
    if err != nil {
        log.Fatal(err)
    }
    ironic_conductor_pxe_init, err := box.FindString("ironic_conductor_pxe_init.sh")
    if err != nil {
        log.Fatal(err)
    }
    ironic_conductor_http, err := box.FindString("ironic_conductor_http.sh")
    if err != nil {
        log.Fatal(err)
    }
    cm := &v1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{
            Name: "ironic-bin",
            Namespace: namespace,
        },
        Data: map[string]string{
            "db-init.py": db_init,
            "db-sync.sh": db_sync,
            "rabbit-init.sh": rabbit_init,
            "ironic-api.sh": ironic_api,
            "ironic-conductor.sh": ironic_conductor,
            "ironic-conductor-init.sh": ironic_conductor_init,
            "ironic-conductor-pxe.sh": ironic_conductor_pxe,
            "ironic-conductor-pxe-init.sh": ironic_conductor_pxe_init,
            "ironic-conductor-http.sh": ironic_conductor_http,
        },
    }
    return cm, nil
}

func GetIronicEtcConfigMap(namespace string, client client.Client) (*v1.ConfigMap, error) {
    // read all bin scripts
    box := packr.New("files", "../../files")

    ironic_conf, err := box.FindString("ironic.conf")
    if err != nil {
        log.Fatal(err)
    }
    policy_json, err := box.FindString("policy.json")
    if err != nil {
        log.Fatal(err)
    }
    tftp_map, err := box.FindString("tftp_map.txt")
    if err != nil {
        log.Fatal(err)
    }
    nginx_conf, err := box.FindString("nginx.conf")
    if err != nil {
        log.Fatal(err)
    }

    // get rabbit secret
    rabbit_secret := &v1.Secret{}
    err = client.Get(context.TODO(), types.NamespacedName{Name: "ironic-rabbitmq-user", Namespace: namespace}, rabbit_secret)
    ironic_conf = strings.Replace(ironic_conf, "##RABBIT_CONNECTION##", string(rabbit_secret.Data["RABBITMQ_TRANSPORT"]), -1)


    // get mysql secret
    mysql_secret := &v1.Secret{}
    err = client.Get(context.TODO(), types.NamespacedName{Name: "ironic-db-user", Namespace: namespace}, mysql_secret)
    mysql_connection_string := fmt.Sprintf("mysql+pymysql://%s:%s@%s:3306/%s?charset=utf8mb4", mysql_secret.Data["DB_USER"],
        mysql_secret.Data["DB_PASSWORD"], mysql_secret.Data["DB_HOST"], mysql_secret.Data["DB_DATABASE"])
    ironic_conf = strings.Replace(ironic_conf, "##MYSQL_CONNECTION##", mysql_connection_string, -1)

    cm := &v1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{
            Name: "ironic-etc",
            Namespace: namespace,
        },
        Data: map[string]string{
            "ironic.conf": ironic_conf,
            "policy.json": policy_json,
            "tftp-map-file": tftp_map,
            "nginx.conf": nginx_conf,
        },
    }
    return cm, nil
}

func GetDHCPConfigMap(namespace string) (*v1.ConfigMap, error) {
    // read all bin scripts
    box := packr.New("files", "../../files")

    dhcp_init, err := box.FindString("dhcp_server_init.sh")
    if err != nil {
        log.Fatal(err)
    }
    dhcp_server, err := box.FindString("dhcp_server.sh")
    if err != nil {
        log.Fatal(err)
    }

    cm := &v1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{
            Name: "dhcp-bin",
            Namespace: namespace,
        },
        Data: map[string]string{
            "dhcp-server-init.sh": dhcp_init,
            "dhcp-server.sh": dhcp_server,
        },
    }

    return cm, nil
}

func GetDHCPEtcConfigMap(namespace string) (*v1.ConfigMap, error) {
    // read all bin scripts
    box := packr.New("files", "../../files")

    dhcp_etc, err := box.FindString("dhcp.conf")
    if err != nil {
        log.Fatal(err)
    }

    cm := &v1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{
            Name: "dhcp-etc",
            Namespace: namespace,
        },
        Data: map[string]string{
            "dhcp-config": dhcp_etc,
        },
    }
    return cm, nil
}

// deploymentForIronic returns a ironic Deployment object
func GetDeploymentForIronic(name string, namespace string, images map[string]string) *appsv1.Deployment {
    ls := GetLabelsForIronic(name)
    var replicas int32 = 1

    var readMode int32 = 0444
    var execMode int32 = 0555
    var rootUser int64 = 0
    var privTrue bool = true

    node_selector := map[string]string{"ironic-control-plane": "enabled"}

    dep := &appsv1.Deployment{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "apps/v1",
            Kind:       "Deployment",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: namespace,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
               MatchLabels: ls,
            },
            Template: v1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: ls,
                },
                Spec: v1.PodSpec{
                    NodeSelector: node_selector,
                    SecurityContext: &v1.PodSecurityContext {
                        RunAsUser: &rootUser,
                    },
                    HostNetwork: true,
                    HostIPC: true,
                    DNSPolicy: "ClusterFirstWithHostNet",
                    InitContainers: []v1.Container{
                        {
                            Image: images["KUBERNETES_ENTRYPOINT"],
                            Name: "init",
                            ImagePullPolicy: "IfNotPresent",
                            Env: []v1.EnvVar{
                                {
                                    Name: "PATH",
                                    Value: "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/",
                                },
                                {
                                    Name: "DEPENDENCY_JOBS_JSON",
                                    Value: fmt.Sprintf("[{'namespace: '%s', 'name': 'ironic-db-sync'}, {'namespace': '%s', 'name': 'ironic-db-init'}]", namespace, namespace),
                                },
                                {
                                    Name: "COMMAND",
                                    Value: "echo done",
                                },
                            },
                            Command: []string{"kubernetes-entrypoint"},
                        },
                        {
                            Name: "ironic-conductor-pxe-init",
                            Image: images["IRONIC_PXE"],
                            ImagePullPolicy: "IfNotPresent",
                            Command: []string { "/tmp/ironic-conductor-pxe-init.sh" },
                            VolumeMounts: []v1.VolumeMount {
                                {
                                    Name: "ironic-bin",
                                    MountPath: "/tmp/ironic-conductor-pxe-init.sh",
                                    SubPath: "ironic-conductor-pxe-init.sh",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "pod-data",
                                    MountPath: "/var/lib/pod_data",
                                },
                            },
                        },
                        {
                            Name: "ironic-conductor-init",
                            Image: images["IRONIC_CONDUCTOR"],
                            ImagePullPolicy: "IfNotPresent",
                            Env: []v1.EnvVar {
                                {
                                    Name: "PXE_NIC",
                                    ValueFrom: &v1.EnvVarSource {
                                        ConfigMapKeyRef: &v1.ConfigMapKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "pxe-settings",
                                            },
                                            Key: "PXE_NIC",
                                        },
                                    },
                                },
                            },
                            VolumeMounts: []v1.VolumeMount {
                                {
                                    Name: "ironic-bin",
                                    MountPath: "/tmp/ironic-conductor-init.sh",
                                    SubPath: "ironic-conductor-init.sh",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "pod-shared",
                                    MountPath: "/tmp/pod-shared",
                                },
                            },
                            Command: []string { "/tmp/ironic-conductor-init.sh" },
                        },
                    },
                    Containers: []v1.Container{
                        {
                            Image:   images["IRONIC_API"],
                            Name:    "ironic-api",
                            ImagePullPolicy: "IfNotPresent",
                            Command: []string{"/tmp/ironic-api.sh", "start"},
                            Lifecycle: &v1.Lifecycle{
                                PreStop: &v1.Handler{
                                    Exec: &v1.ExecAction{
                                        Command: []string{"/tmp/ironic-api.sh", "stop"},
                                    },
                                },
                            },
                            Ports: []v1.ContainerPort{
                                {
                                    ContainerPort: 6385,
                                },
                            },
                            ReadinessProbe: &v1.Probe{
                                Handler: v1.Handler{
                                    TCPSocket: &v1.TCPSocketAction{
                                        Port: intstr.FromInt(6385),
                                    },
                                },
                            },
                            VolumeMounts: []v1.VolumeMount{
                                {
                                    Name: "ironic-bin",
                                    MountPath: "/tmp/ironic-api.sh",
                                    SubPath: "ironic-api.sh",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "ironic-etc",
                                    MountPath: "/etc/ironic/ironic.conf",
                                    SubPath: "ironic.conf",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "ironic-etc",
                                    MountPath: "/etc/ironic/logging.conf",
                                    SubPath: "logging.conf",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "ironic-etc",
                                    MountPath: "/etc/ironic/policy.json",
                                    SubPath: "policy.json",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "pod-shared",
                                    MountPath: "/tmp/pod-shared",
                                },
                            },
                        },
                        {
                            Name: "ironic-conductor",
                            Image: images["IRONIC_CONDUCTOR"],
                            ImagePullPolicy: "IfNotPresent",
                            SecurityContext: &v1.SecurityContext {
                                Privileged: &privTrue,
                                RunAsUser: &rootUser,
                            },
                            Command: []string { "/tmp/ironic-conductor.sh" },
                            VolumeMounts: []v1.VolumeMount {
                                {
                                    Name: "ironic-bin",
                                    MountPath: "/tmp/ironic-conductor.sh",
                                    SubPath: "ironic-conductor.sh",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "pod-shared",
                                    MountPath: "/tmp/pod-shared",
                                },
                                {
                                    Name: "pod-var-cache-ironic",
                                    MountPath: "/var/cache/ironic",
                                },
                                {
                                    Name: "ironic-etc",
                                    MountPath: "/etc/ironic/ironic.conf",
                                    SubPath: "ironic.conf",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "ironic-etc",
                                    MountPath: "/etc/ironic/logging.conf",
                                    SubPath: "logging.conf",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "pod-data",
                                    MountPath: "/var/lib/pod_data",
                                },
                            },
                        },
                        {
                            Name: "ironic-conductor-pxe",
                            Image: images["IRONIC_PXE"],
                            ImagePullPolicy: "IfNotPresent",
                            SecurityContext: &v1.SecurityContext {
                                Privileged: &privTrue,
                                RunAsUser: &rootUser,
                            },
                            Env: []v1.EnvVar {
                                {
                                    Name: "PXE_NIC",
                                    ValueFrom: &v1.EnvVarSource {
                                        ConfigMapKeyRef: &v1.ConfigMapKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "pxe-settings",
                                            },
                                            Key: "PXE_NIC",
                                        },
                                    },
                               },
                            },
                            Command: []string { "/tmp/ironic-conductor-pxe.sh" },
                            VolumeMounts: []v1.VolumeMount {
                                {
                                    Name: "ironic-bin",
                                    MountPath: "/tmp/ironic-conductor-pxe.sh",
                                    SubPath: "ironic-conductor-pxe.sh",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "ironic-etc",
                                    MountPath: "/tftp-map-file",
                                    SubPath: "tftp-map-file",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "pod-data",
                                    MountPath: "/var/lib/pod_data",
                                },
                            },
                            Ports: []v1.ContainerPort {
                                {
                                    ContainerPort: 69,
                                    HostPort: 69,
                                    Protocol: "UDP",
                                },
                            },
                        },
                        {
                            Name: "ironic-conductor-http",
                            Image: images["NGINX"],
                            ImagePullPolicy: "IfNotPresent",
                            Command: []string { "/tmp/ironic-conductor-http.sh" },
                            VolumeMounts: []v1.VolumeMount {
                                {
                                    Name: "ironic-bin",
                                    MountPath: "/tmp/ironic-conductor-http.sh",
                                    SubPath: "ironic-conductor-http.sh",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "ironic-etc",
                                    MountPath: "/etc/nginx/nginx.conf",
                                    SubPath: "nginx.conf",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "pod-data",
                                    MountPath: "/var/lib/pod_data",
                                },
                            },
                            Ports: []v1.ContainerPort {
                                {
                                    ContainerPort: 8081,
                                    HostPort: 8081,
                                    Protocol: "TCP",
                                },
                            },
                        },
                    },
                    Volumes: []v1.Volume{
                        {
                            Name: "ironic-bin",
                            VolumeSource: v1.VolumeSource {
                                ConfigMap: &v1.ConfigMapVolumeSource {
                                    DefaultMode: &execMode,
                                    LocalObjectReference: v1.LocalObjectReference{
                                        Name: "ironic-bin",
                                    },
                                },
                            },
                        },
                        {
                            Name: "ironic-etc",
                            VolumeSource: v1.VolumeSource {
                                ConfigMap: &v1.ConfigMapVolumeSource {
                                    DefaultMode: &readMode,
                                    LocalObjectReference: v1.LocalObjectReference{
                                        Name: "ironic-etc",
                                    },
                                },
                            },
                        },
                        {
                            Name: "pod-shared",
                            VolumeSource: v1.VolumeSource {
                                EmptyDir: &v1.EmptyDirVolumeSource {},
                            },
                        },
                        {
                            Name: "pod-data",
                            VolumeSource: v1.VolumeSource {
                                EmptyDir: &v1.EmptyDirVolumeSource {},
                            },
                        },
                        {
                            Name: "pod-var-cache-ironic",
                            VolumeSource: v1.VolumeSource {
                                EmptyDir: &v1.EmptyDirVolumeSource {},
                            },
                        },
                        {
                            Name: "ironic-bin",
                            VolumeSource: v1.VolumeSource {
                                ConfigMap: &v1.ConfigMapVolumeSource {
                                    DefaultMode: &execMode,
                                    LocalObjectReference: v1.LocalObjectReference {
                                        Name: "ironic-bin",
                                    },
                                },
                            },
                        },
                        {
                            Name: "ironic-etc",
                            VolumeSource: v1.VolumeSource {
                                ConfigMap: &v1.ConfigMapVolumeSource {
                                    DefaultMode: &readMode,
                                    LocalObjectReference: v1.LocalObjectReference {
                                        Name: "ironic-etc",
                                    },
                                },
                            },
                        },
                    },
                },
            },
       },
    }
    return dep
}

// GetLabelsForIronic returns the labels for selecting the resources
// belonging to the given ironic CR name.
func GetLabelsForIronic(name string) map[string]string {
        return map[string]string{"app": "ironic", "ironic_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func GetPodNames(pods []v1.Pod) []string {
        var podNames []string
        for _, pod := range pods {
                podNames = append(podNames, pod.Name)
        }
        return podNames
}

// serviceForIronicApi returns a ironic-api Service object
func GetServiceForIronicApi(name string, namespace string) *v1.Service {
    srv_selector := map[string]string{"app": "ironic", "ironic_cr": "openstack-ironicapi"}
    srv := &v1.Service{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "core/v1",
            Kind:       "Service",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: namespace,
        },
        Spec: v1.ServiceSpec{
            Type: "NodePort",
            Selector: srv_selector,
            Ports: []v1.ServicePort{
                {
                    Name: "ironic-api",
                    Protocol: "TCP",
                    Port: 6385,
                    NodePort: 32732,
                },
            },
        },
    }
    return srv
}

func GetDbInitJob(namespace string, images map[string]string) *batchv1.Job {
    node_selector := map[string]string{"ironic-control-plane": "enabled"}
    var readMode int32 = 0444
    var execMode int32 = 0555

    job := &batchv1.Job{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "batch/v1",
            Kind:       "Job",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name:      "ironic-db-init",
            Namespace: namespace,
        },
        Spec: batchv1.JobSpec {
            Template: v1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta {
                    Labels: map[string]string {"app": "ironic", "ironicapi_cr": "openstack-ironicapi", "component": "db-init" },
                },
                Spec: v1.PodSpec {
                    NodeSelector: node_selector,
                    RestartPolicy: "OnFailure",
                    Containers: []v1.Container {
                        {
                            Name: "ironic-db-init-0",
                            Image: images["IRONIC_API"],
                            ImagePullPolicy: "IfNotPresent",
                            Env: []v1.EnvVar {
                                {
                                    Name: "ROOT_DB_HOST",
                                    ValueFrom: &v1.EnvVarSource {
                                        SecretKeyRef: &v1.SecretKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "mysql-root-credentials",
                                            },
                                            Key: "ROOT_DB_HOST",
                                        },
                                    },
                                },
                                {
                                    Name: "ROOT_DB_USER",
                                    ValueFrom: &v1.EnvVarSource {
                                        SecretKeyRef: &v1.SecretKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "mysql-root-credentials",
                                            },
                                            Key: "ROOT_DB_USER",
                                        },
                                    },
                                },
                                {
                                    Name: "ROOT_DB_PASSWORD",
                                    ValueFrom: &v1.EnvVarSource {
                                        SecretKeyRef: &v1.SecretKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "mysql-root-credentials",
                                            },
                                            Key: "ROOT_DB_PASSWORD",
                                        },
                                    },
                                },
                                {
                                    Name: "USER_DB_HOST",
                                    ValueFrom: &v1.EnvVarSource {
                                        SecretKeyRef: &v1.SecretKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "ironic-db-user",
                                            },
                                            Key: "DB_HOST",
                                        },
                                    },
                                },
                                {
                                    Name: "USER_DB_USER",
                                    ValueFrom: &v1.EnvVarSource {
                                        SecretKeyRef: &v1.SecretKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "ironic-db-user",
                                            },
                                            Key: "DB_USER",
                                        },
                                    },
                                },
                                {
                                    Name: "USER_DB_PASSWORD",
                                    ValueFrom: &v1.EnvVarSource {
                                        SecretKeyRef: &v1.SecretKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "ironic-db-user",
                                            },
                                            Key: "DB_PASSWORD",
                                        },
                                    },
                                },
                                {
                                    Name: "USER_DB_DATABASE",
                                    ValueFrom: &v1.EnvVarSource {
                                        SecretKeyRef: &v1.SecretKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "ironic-db-user",
                                            },
                                            Key: "DB_DATABASE",
                                        },
                                    },
                                },

                            },
                            Command: []string { "/tmp/db-init.py" },
                            VolumeMounts: []v1.VolumeMount {
                                {
                                    Name: "db-init-py",
                                    MountPath: "/tmp/db-init.py",
                                    SubPath: "db-init.py",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "etc-service",
                                    MountPath: "/etc/ironic",
                                },
                                {
                                    Name: "db-init-conf",
                                    MountPath: "/etc/ironic/ironic.conf",
                                    SubPath: "ironic.conf",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "db-init-conf",
                                    MountPath: "/etc/ironic/logging.conf",
                                    SubPath: "logging.conf",
                                    ReadOnly: true,
                                },
                            },
                        },
                    },
                    Volumes: []v1.Volume {
                        {
                            Name: "etc-service",
                            VolumeSource: v1.VolumeSource {
                                EmptyDir: &v1.EmptyDirVolumeSource {},
                            },
                        },
                        {
                            Name: "db-init-py",
                            VolumeSource: v1.VolumeSource {
                                ConfigMap: &v1.ConfigMapVolumeSource {
                                    DefaultMode: &execMode,
                                    LocalObjectReference: v1.LocalObjectReference {
                                        Name: "ironic-bin",
                                    },
                                },
                            },
                        },
                        {
                            Name: "db-init-conf",
                            VolumeSource: v1.VolumeSource {
                                ConfigMap: &v1.ConfigMapVolumeSource {
                                    DefaultMode: &readMode,
                                    LocalObjectReference: v1.LocalObjectReference {
                                        Name: "ironic-etc",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    return job
}

func GetDbSyncJob(namespace string, images map[string]string) *batchv1.Job {
    node_selector := map[string]string{"ironic-control-plane": "enabled"}
    var readMode int32 = 0444
    var execMode int32 = 0555

    job := &batchv1.Job{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "batch/v1",
            Kind:       "Job",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name:      "ironic-db-sync",
            Namespace: namespace,
        },
        Spec: batchv1.JobSpec {
            Template: v1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta {
                    Labels: map[string]string {"app": "ironic", "ironicapi_cr": "openstack-ironicapi", "component": "db-sync" },
                },
                Spec: v1.PodSpec {
                    NodeSelector: node_selector,
                    RestartPolicy: "OnFailure",
                    InitContainers: []v1.Container {
                        {
                            Name: "init",
                            Image: images["KUBERNETES_ENTRYPOINT"],
                            ImagePullPolicy: "IfNotPresent",
                            Env: []v1.EnvVar {
                                {
                                    Name: "PATH",
                                    Value: "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/",
                                },
                                {
                                    Name: "DEPENDENCY_JOBS_JSON",
                                    Value: fmt.Sprintf("[{'namespace': '%s', 'name': 'ironic-db-init'}]", namespace),
                                },
                                {
                                    Name: "COMMAND",
                                    Value: "echo done",
                                },
                            },
                            Command: []string { "kubernetes-entrypoint" },
                        },
                    },
                    Containers: []v1.Container {
                        {
                            Name: "ironic-db-sync",
                            Image: images["IRONIC_API"],
                            ImagePullPolicy: "IfNotPresent",
                            Command: []string { "/tmp/db-sync.sh" },
                            VolumeMounts: []v1.VolumeMount {
                                {
                                    Name: "db-sync-sh",
                                    MountPath: "/tmp/db-sync.sh",
                                    SubPath: "db-sync.sh",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "etc-service",
                                    MountPath: "/etc/ironic",
                                },
                                {
                                    Name: "db-sync-conf",
                                    MountPath: "/etc/ironic/ironic.conf",
                                    SubPath: "ironic.conf",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "db-sync-conf",
                                    MountPath: "/etc/ironic/logging.conf",
                                    SubPath: "logging.conf",
                                    ReadOnly: true,
                                },
                            },
                        },
                    },
                    Volumes: []v1.Volume {
                        {
                            Name: "etc-service",
                            VolumeSource: v1.VolumeSource {
                                EmptyDir: &v1.EmptyDirVolumeSource {},
                            },
                        },
                        {
                            Name: "db-sync-sh",
                            VolumeSource: v1.VolumeSource {
                                ConfigMap: &v1.ConfigMapVolumeSource {
                                    DefaultMode: &execMode,
                                    LocalObjectReference: v1.LocalObjectReference {
                                        Name: "ironic-bin",
                                    },
                                },
                            },
                        },
                        {
                            Name: "db-sync-conf",
                            VolumeSource: v1.VolumeSource {
                                ConfigMap: &v1.ConfigMapVolumeSource {
                                    DefaultMode: &readMode,
                                    LocalObjectReference: v1.LocalObjectReference {
                                        Name: "ironic-etc",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    return job
}

func GetDHCPService(namespace string) *v1.Service {
    selector := map[string]string{"app": "dhcp-server"}

    service := &v1.Service{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "v1",
            Kind:       "Service",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name:      "dhcp-server",
            Namespace: namespace,
        },
        Spec: v1.ServiceSpec {
            Type: "ClusterIP",
            Ports: []v1.ServicePort {
                {
                    Name: "dhcp",
                    Port: 67,
                    Protocol: "UDP",
                    TargetPort: intstr.FromInt(67),
                },
            },
            Selector: selector,
        },
    }

    return service
}
func GetRabbitInitJob(namespace string, images map[string]string) *batchv1.Job {
    node_selector := map[string]string{"ironic-control-plane": "enabled"}
    var execMode int32 = 0555

    job := &batchv1.Job{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "batch/v1",
            Kind:       "Job",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name:      "ironic-rabbit-init",
            Namespace: namespace,
        },
        Spec: batchv1.JobSpec {
            Template: v1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta {
                    Labels: map[string]string {"app": "ironic", "ironicapi_cr": "openstack-ironicapi", "component": "rabbit-init" },
                },
                Spec: v1.PodSpec {
                    NodeSelector: node_selector,
                    RestartPolicy: "OnFailure",
                    Containers: []v1.Container {
                        {
                            Name: "rabbit-init",
                            Image: images["RABBIT_MANAGEMENT"],
                            ImagePullPolicy: "IfNotPresent",
                            Env: []v1.EnvVar {
                                {
                                    Name: "RABBITMQ_ADMIN_CONNECTION",
                                    ValueFrom: &v1.EnvVarSource {
                                        SecretKeyRef: &v1.SecretKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "ironic-rabbitmq-admin",
                                            },
                                            Key: "RABBITMQ_CONNECTION",
                                        },
                                    },
                                },
                                {
                                    Name: "RABBITMQ_USER_CONNECTION",
                                    ValueFrom: &v1.EnvVarSource {
                                        SecretKeyRef: &v1.SecretKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "ironic-rabbitmq-user",
                                            },
                                            Key: "RABBITMQ_CONNECTION",
                                        },
                                    },
                                },
                            },
                            Command: []string { "/tmp/rabbit-init.sh" },
                            VolumeMounts: []v1.VolumeMount {
                                {
                                    Name: "rabbit-init-sh",
                                    MountPath: "/tmp/rabbit-init.sh",
                                    SubPath: "rabbit-init.sh",
                                    ReadOnly: true,
                                },
                            },
                        },
                    },
                    Volumes: []v1.Volume {

                        {
                            Name: "rabbit-init-sh",
                            VolumeSource: v1.VolumeSource {
                                ConfigMap: &v1.ConfigMapVolumeSource {
                                    DefaultMode: &execMode,
                                    LocalObjectReference: v1.LocalObjectReference {
                                        Name: "ironic-bin",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    return job
}

func GetDHCPDeployment(namespace string, images map[string]string) *appsv1.Deployment {
    label_selector := map[string]string{"apps": "dhcp-server"}
    var replicas int32 = 1
    var readMode int32 = 0444
    var execMode int32 = 0555

    dep := &appsv1.Deployment {
        TypeMeta: metav1.TypeMeta{
            APIVersion: "apps/v1",
            Kind:       "Deployment",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name:      "dhcp-server",
            Namespace: namespace,
        },
        Spec: appsv1.DeploymentSpec {
            Replicas: &replicas,
            Selector: &metav1.LabelSelector {
                MatchLabels: label_selector,
            },
            Template: v1.PodTemplateSpec {
                ObjectMeta: metav1.ObjectMeta {
                    Labels: label_selector,
                },
                Spec: v1.PodSpec {
                    HostNetwork: true,
                    InitContainers: []v1.Container {
                        {
                            Name: "init-dhcp",
                            Image: images["IRONIC_PXE"],
                            ImagePullPolicy: "IfNotPresent",
                            Command: []string {"/tmp/scripts/dhcp-server-init.sh"},
                            Env: []v1.EnvVar {
                                {
                                    Name: "PXE_NIC",
                                    ValueFrom: &v1.EnvVarSource {
                                        ConfigMapKeyRef: &v1.ConfigMapKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "pxe-settings",
                                            },
                                            Key: "PXE_NIC",
                                        },
                                    },
                                },
                                {
                                    Name: "DHCP_HOSTS",
                                    ValueFrom: &v1.EnvVarSource {
                                        ConfigMapKeyRef: &v1.ConfigMapKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "dhcp-settings",
                                            },
                                            Key: "DHCP_HOSTS",
                                        },
                                    },
                                },
                                {
                                    Name: "CLUSTER_DOMAIN",
                                    ValueFrom: &v1.EnvVarSource {
                                        ConfigMapKeyRef: &v1.ConfigMapKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "dhcp-settings",
                                            },
                                            Key: "CLUSTER_DOMAIN",
                                        },
                                    },
                                },
                                {
                                    Name: "INITIAL_IP_RANGE",
                                    ValueFrom: &v1.EnvVarSource {
                                        ConfigMapKeyRef: &v1.ConfigMapKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "dhcp-settings",
                                            },
                                            Key: "INITIAL_IP_RANGE",
                                        },
                                    },
                                },
                                {
                                    Name: "FINAL_IP_RANGE",
                                    ValueFrom: &v1.EnvVarSource {
                                        ConfigMapKeyRef: &v1.ConfigMapKeySelector {
                                            LocalObjectReference: v1.LocalObjectReference {
                                                Name: "dhcp-settings",
                                            },
                                            Key: "FINAL_IP_RANGE",
                                        },
                                    },
                                },
                            },
                            VolumeMounts: []v1.VolumeMount {
                                {
                                    Name: "dhcp-bin",
                                    MountPath: "/tmp/scripts/",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "dhcp-hosts",
                                    MountPath: "/data/hosts/",
                                },
                                {
                                    Name: "dhcp-zones",
                                    MountPath: "/data/zones/",
                                },
                            },
                        },
                    },
                    Containers: []v1.Container {
                        {
                            Name: "dhcp-server",
                            Image: images["IRONIC_PXE"],
                            ImagePullPolicy: "IfNotPresent",
                            Command: []string { "/tmp/scripts/dhcp-server.sh" },
                            Ports: []v1.ContainerPort {
                                {
                                    ContainerPort: 67,
                                    HostPort: 67,
                                    Protocol: "UDP",
                                },
                            },
                            VolumeMounts: []v1.VolumeMount {
                                {
                                    Name: "dhcp-bin",
                                    MountPath: "/tmp/scripts/",
                                },
                                {
                                    Name: "dhcp-etc",
                                    MountPath: "/data/dhcpd.conf",
                                    SubPath: "dhcpd.conf",
                                },
                                {
                                    Name: "dhcp-zones",
                                    MountPath: "/data/zones/",
                                },
                                {
                                    Name: "dhcp-hosts",
                                    MountPath: "/data/hosts/",
                                },
                            },
                        },
                    },
                    Volumes: []v1.Volume {
                        {
                            Name: "dhcp-bin",
                            VolumeSource: v1.VolumeSource {
                                ConfigMap: &v1.ConfigMapVolumeSource {
                                    DefaultMode: &execMode,
                                    LocalObjectReference: v1.LocalObjectReference {
                                        Name: "dhcp-bin",
                                    },
                                },
                            },
                        },
                        {
                            Name: "dhcp-etc",
                            VolumeSource: v1.VolumeSource {
                                ConfigMap: &v1.ConfigMapVolumeSource {
                                    DefaultMode: &readMode,
                                    Items: []v1.KeyToPath {
                                        {
                                           Key: "dhcp-config",
                                           Path: "dhcpd.conf",
                                       },
                                    },
                                    LocalObjectReference: v1.LocalObjectReference {
                                        Name: "dhcp-etc",
                                    },
                                },
                            },
                        },
                        {
                            Name: "dhcp-hosts",
                            VolumeSource: v1.VolumeSource {
                                EmptyDir: &v1.EmptyDirVolumeSource {},
                            },
                        },
                        {
                            Name: "dhcp-zones",
                            VolumeSource: v1.VolumeSource {
                                EmptyDir: &v1.EmptyDirVolumeSource {},
                            },
                        },
                    },
                },
            },
        },
    }

    return dep
}


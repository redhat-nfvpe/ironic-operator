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

    v1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/types"
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
    ironic_conf = strings.Replace(ironic_conf, "##RABBIT_CONNECTION##", string(rabbit_secret.Data["RABBITMQ_CONNECTION"]), -1)

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

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
    "io/ioutil"
    "log"
    ironicv1alpha1 "github.com/redhat-nfvpe/ironic-operator/pkg/apis/ironic/v1alpha1"

    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetIronicBinConfigMap(m *ironicv1alpha1.IronicApi) (*v1.ConfigMap, error) {
    // read all bin scripts
    db_init, err := ioutil.ReadFile("files/db_init.py")
    if err != nil {
        log.Fatal(err)
    }
    db_sync, err := ioutil.ReadFile("files/db_sync.sh")
    if err != nil {
        log.Fatal(err)
    }
    rabbit_init, err := ioutil.ReadFile("files/rabbit_init.sh")
    if err != nil {
        log.Fatal(err)
    }
    ironic_api, err := ioutil.ReadFile("files/ironic_api.sh")
    if err != nil {
        log.Fatal(err)
    }
    ironic_conductor, err := ioutil.ReadFile("files/ironic_conductor.sh")
    if err != nil {
        log.Fatal(err)
    }
    ironic_conductor_init, err := ioutil.ReadFile("files/ironic_conductor_init.sh")
    if err != nil {
        log.Fatal(err)
    }
    ironic_conductor_pxe, err := ioutil.ReadFile("files/ironic_conductor_pxe.sh")
    if err != nil {
        log.Fatal(err)
    }
    ironic_conductor_pxe_init, err := ioutil.ReadFile("files/ironic_conductor_pxe_init.sh")
    if err != nil {
        log.Fatal(err)
    }
    ironic_conductor_http, err := ioutil.ReadFile("files/ironic_conductor_http.sh")
    if err != nil {
        log.Fatal(err)
    }
    ironic_conductor_http_init, err := ioutil.ReadFile("files/ironic_conductor_http_init.sh")
    if err != nil {
        log.Fatal(err)
    }
    cm := &v1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{
            Name: "ironic-bin",
            Namespace: m.Namespace,
        },
        Data: map[string]string{
            "db-init.py": string(db_init),
            "db-sync.sh": string(db_sync),
            "rabbit-init.sh": string(rabbit_init),
            "ironic-api.sh": string(ironic_api),
            "ironic-conductor.sh": string(ironic_conductor),
            "ironic-conductor-init.sh": string(ironic_conductor_init),
            "ironic-conductor-pxe.sh": string(ironic_conductor_pxe),
            "ironic-conductor-pxe-init.sh": string(ironic_conductor_pxe_init),
            "ironic-conductor-http.sh": string(ironic_conductor_http),
            "ironic-conductor-http-init.sh": string(ironic_conductor_http_init),
        },
    }
    return cm, nil
}

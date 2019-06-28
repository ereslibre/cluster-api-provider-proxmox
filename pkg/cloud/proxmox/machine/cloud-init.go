/*
 * Copyright 2019 Rafael Fernández López <ereslibre@ereslibre.es>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package machine

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"

	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

var (
	cloudInitDataControlPlane = `#cloud-config
hostname: {{.Hostname}}
write_files:
- path: /tmp/kubeadm.conf
  owner: root:root
  permissions: '0600'
  content: |
    apiVersion: kubeadm.k8s.io/v1beta1
    kind: InitConfiguration
    bootstrapTokens:
      - token: g0mea5.sh7485cyaz49mmse
    ---
    apiVersion: kubeadm.k8s.io/v1beta1
    kind: ClusterConfiguration
    kubernetesVersion: v1.14.3
packages:
  - qemu-guest-agent
  - policykit-1
  - docker.io
runcmd:
  - systemctl enable --now qemu-guest-agent
  - systemctl enable --now docker
  - apt-get update
  - apt-get install -y apt-transport-https curl
  - bash -c "curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -"
  - echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list
  - apt-get update
  - apt-get install -y kubelet kubeadm kubectl
  - apt-mark hold kubelet kubeadm kubectl
  - kubeadm init --config /tmp/kubeadm.conf
`

	cloudInitDataWorker = `#cloud-config
hostname: {{.Hostname}}
write_files:
- path: /tmp/kubeadm.conf
  owner: root:root
  permissions: '0600'
  content: |
    apiVersion: kubeadm.k8s.io/v1beta1
    kind: JoinConfiguration
    discovery:
      bootstrapToken:
        apiServerEndpoint: {{.ControlPlaneEndpoint}}
        token: g0mea5.sh7485cyaz49mmse
        unsafeSkipCAVerification: true
packages:
  - qemu-guest-agent
  - policykit-1
  - docker.io
runcmd:
  - systemctl enable --now qemu-guest-agent
  - systemctl enable --now docker
  - apt-get update
  - apt-get install -y apt-transport-https curl
  - bash -c "curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -"
  - echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list
  - apt-get update
  - apt-get install -y kubelet kubeadm kubectl
  - apt-mark hold kubelet kubeadm kubectl
  - kubeadm join --config /tmp/kubeadm.conf
`

	cloudInitNetworkData = `version: 2
ethernets:
  interfaces:
    match:
      name: ens*
    dhcp4: yes
`
)

func (a *Actuator) cloudInitUserDataForMachine(cluster *clusterv1alpha1.Cluster, machine *clusterv1alpha1.Machine) (string, error) {
	switch machine.Labels["set"] {
	case "controlplane":
		tmpl, err := template.New("node-cloud-init-config").Parse(cloudInitDataControlPlane)
		if err != nil {
			return "", err
		}
		var rendered bytes.Buffer
		err = tmpl.Execute(&rendered, struct {
			Hostname string
		}{
			Hostname: machine.ObjectMeta.Name,
		})
		if err != nil {
			return "", err
		}
		return rendered.String(), nil
	case "node":
		tmpl, err := template.New("node-cloud-init-config").Parse(cloudInitDataWorker)
		if err != nil {
			return "", err
		}
		controlPlaneEndpoint, err := a.getControlPlaneEndpoint(cluster)
		if err != nil {
			return "", err
		}
		var rendered bytes.Buffer
		err = tmpl.Execute(&rendered, struct {
			Hostname             string
			ControlPlaneEndpoint string
		}{
			Hostname:             machine.ObjectMeta.Name,
			ControlPlaneEndpoint: controlPlaneEndpoint,
		})
		if err != nil {
			return "", err
		}
		return rendered.String(), nil
	}
	return "", errors.Errorf("unknown machine role %q for machine %q", machine.Labels["set"], machine.ObjectMeta.Name)
}

func (a *Actuator) cloudInitNetworkForMachine(cluster *clusterv1alpha1.Cluster, machine *clusterv1alpha1.Machine) (string, error) {
	return cloudInitNetworkData, nil
}

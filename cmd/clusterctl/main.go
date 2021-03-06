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

package main

import (
	"github.com/ereslibre/cluster-api-provider-proxmox/pkg/cloud/proxmox"
	"github.com/ereslibre/cluster-api-provider-proxmox/pkg/cloud/proxmox/machine"
	"k8s.io/klog"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/cmd"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
)

func main() {
	var err error
	machine.MachineActuator, err = machine.NewActuator(proxmox.ActuatorParams{})
	if err != nil {
		klog.Fatalf("Error creating cluster provisioner for proxmox: %v", err)
	}
	common.RegisterClusterProvisioner("proxmox", machine.MachineActuator)
	cmd.Execute()
}

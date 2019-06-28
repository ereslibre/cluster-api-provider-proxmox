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
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	proxmoxapi "github.com/ereslibre/proxmox-api-go/proxmox"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"

	kubernetesclient "k8s.io/client-go/kubernetes"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	clusterapi "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
	"sigs.k8s.io/cluster-api/pkg/util"

	"github.com/ereslibre/cluster-api-provider-proxmox/pkg/cloud/proxmox"
)

const (
	vmIdAnnotationKey = "cluster-api-provider-proxmox/vmid"

	createCheckPeriod  = 10 * time.Second
	createCheckTimeout = 5 * time.Minute
)

type Actuator struct {
	client        *proxmoxapi.Client
	clusterClient clusterapi.Interface
	kubeClient    kubernetes.Interface
	crClient      crclient.Client
}

var MachineActuator *Actuator

func NewActuator(params proxmox.ActuatorParams) (*Actuator, error) {
	tlsconf := tls.Config{
		InsecureSkipVerify: true,
	}
	client, err := proxmoxapi.NewClient(fmt.Sprintf("https://%s/api2/json", os.Getenv("PROXMOX_HOSTPORT")), nil, &tlsconf)
	if err != nil {
		return nil, err
	}
	if err := client.Login(os.Getenv("PROXMOX_USERNAME"), os.Getenv("PROXMOX_PASSWORD")); err != nil {
		return nil, err
	}
	return &Actuator{
		client:        client,
		clusterClient: params.ClusterClient,
		kubeClient:    params.KubeClient,
		crClient:      params.Client,
	}, nil
}

func proxmoxHypervisorName() string {
	return os.Getenv("PROXMOX_HYPERVISOR_NAME")
}

func ciTemplatesStorage() string {
	return os.Getenv("PROXMOX_HYPERVISOR_SNIPPETS_STORAGE")
}

func newVmRef(vmId, node string) (*proxmoxapi.VmRef, error) {
	vmId_, err := strconv.Atoi(vmId)
	if err != nil {
		return nil, err
	}
	vmRef := proxmoxapi.NewVmRef(vmId_)
	vmRef.SetNode(node)
	return vmRef, nil
}

func (a *Actuator) getVMRef(machine *clusterv1alpha1.Machine) (*proxmoxapi.VmRef, error) {
	vmRef, err := newVmRef(machine.ObjectMeta.Annotations[vmIdAnnotationKey], proxmoxHypervisorName())
	if err != nil {
		return nil, err
	}
	return vmRef, nil
}

func (a *Actuator) getVmInfo(machine *clusterv1alpha1.Machine) (map[string]interface{}, error) {
	if vmId, ok := machine.ObjectMeta.Annotations[vmIdAnnotationKey]; ok {
		vmRef, err := newVmRef(vmId, proxmoxHypervisorName())
		if err != nil {
			return nil, err
		}
		vmInfo, err := a.client.GetVmInfo(vmRef)
		return vmInfo, err
	}
	return nil, fmt.Errorf("getVmInfo: Vm %q not found", machine.ObjectMeta.Name)
}

func (a *Actuator) uploadCiConfigForMachine(cluster *clusterv1alpha1.Cluster, machine *clusterv1alpha1.Machine) error {
	cloudInitData, err := a.cloudInitUserDataForMachine(cluster, machine)
	if err != nil {
		return err
	}
	cloudInitNetwork, err := a.cloudInitNetworkForMachine(cluster, machine)
	if err != nil {
		return err
	}

	err = a.client.Upload(proxmoxapi.StorageRef{
		Node:    proxmoxHypervisorName(),
		Storage: ciTemplatesStorage(),
	}, fmt.Sprintf("%s-userdata.cfg", machine.ObjectMeta.Name), cloudInitData)
	if err != nil {
		return err
	}

	err = a.client.Upload(proxmoxapi.StorageRef{
		Node:    proxmoxHypervisorName(),
		Storage: ciTemplatesStorage(),
	}, fmt.Sprintf("%s-network.cfg", machine.ObjectMeta.Name), cloudInitNetwork)

	return err
}

func (a *Actuator) cloneTemplate(vmRef *proxmoxapi.VmRef, name string) error {
	templateId, err := strconv.Atoi(os.Getenv("VM_TEMPLATE_ID"))
	if err != nil {
		return err
	}
	templateRef := proxmoxapi.NewVmRef(templateId)
	templateRef.SetNode(proxmoxHypervisorName())
	cloneParams := map[string]interface{}{
		"newid":  vmRef.VmId(),
		"target": proxmoxHypervisorName(),
		"name":   name,
		"full":   false,
	}
	_, err = a.client.CloneQemuVm(templateRef, cloneParams)
	if err != nil {
		return err
	}
	config, err := a.client.GetVmConfig(vmRef)
	if err != nil {
		return err
	}
	config["cicustom"] = fmt.Sprintf("user=%s:snippets/%s-userdata.cfg,network=%s:snippets/%s-network.cfg", ciTemplatesStorage(), name, ciTemplatesStorage(), name)
	config["agent"] = "enabled=1"
	config["cores"] = "2"
	_, err = a.client.SetVmConfigSync(vmRef, config)
	return err
}

func (a *Actuator) resizeDisk(vmRef *proxmoxapi.VmRef, disk string, extraGB int) error {
	_, err := a.client.ResizeQemuDisk(vmRef, disk, extraGB)
	return err
}

func (a *Actuator) getControlPlaneEndpoint(cluster *clusterv1alpha1.Cluster) (string, error) {
	if len(cluster.Status.APIEndpoints) == 0 {
		return "", fmt.Errorf("master endpoint not found in apiEndpoints for cluster %q", cluster.ObjectMeta.Name)
	}
	apiEndpoint := cluster.Status.APIEndpoints[0]
	return fmt.Sprintf("%s:%d", apiEndpoint.Host, apiEndpoint.Port), nil
}

func (a *Actuator) Create(ctx context.Context, cluster *clusterv1alpha1.Cluster, machine *clusterv1alpha1.Machine) error {
	klog.Info("[proxmox] Creating machine")

	machine, _ = a.setMachinePhase(machine, "Pending")

	if err := a.uploadCiConfigForMachine(cluster, machine); err != nil {
		return err
	}

	vmNextId, err := a.client.GetNextID(0)
	if err != nil {
		return err
	}
	vmRef := proxmoxapi.NewVmRef(vmNextId)

	if err := a.cloneTemplate(vmRef, machine.ObjectMeta.Name); err != nil {
		return err
	}

	if err := a.resizeDisk(vmRef, "scsi0", 30); err != nil {
		return err
	}

	if _, err = a.client.StartVm(vmRef); err != nil {
		return err
	}

	if machine.ObjectMeta.Annotations == nil {
		machine.ObjectMeta.Annotations = map[string]string{}
	}
	machine.ObjectMeta.Annotations[vmIdAnnotationKey] = strconv.Itoa(vmNextId)
	if machine.Spec.ProviderID == nil || *machine.Spec.ProviderID == "" {
		providerID := fmt.Sprintf("proxmox://%d", vmNextId)
		machine.Spec.ProviderID = &providerID
	}
	if err := a.crClient.Update(ctx, machine); err != nil {
		return err
	}

	machine, _ = a.setMachinePhase(machine, "Running")

	if a.isControlPlane(machine) {
		err = util.PollImmediate(createCheckPeriod, createCheckTimeout, func() (bool, error) {
			ipAddress, err := a.GetIP(cluster, machine)
			if err != nil {
				return false, nil
			}
			return len(ipAddress) > 0, nil
		})
		if err != nil {
			return err
		}
		if err := a.annotateClusterWithControlPlaneEndpoints(cluster, machine); err != nil {
			return err
		}
	}

	return nil
}

func (a *Actuator) annotateClusterWithControlPlaneEndpoints(cluster *clusterv1alpha1.Cluster, machine *clusterv1alpha1.Machine) error {
	if cluster.Status.APIEndpoints == nil {
		cluster.Status.APIEndpoints = []clusterv1alpha1.APIEndpoint{}
	}
	ipAddress, err := a.GetIP(cluster, machine)
	if err != nil {
		return err
	}
	cluster.Status.APIEndpoints = append(cluster.Status.APIEndpoints, clusterv1alpha1.APIEndpoint{
		Host: ipAddress,
		Port: 6443,
	})
	_, err = a.clusterClient.ClusterV1alpha1().Clusters(cluster.ObjectMeta.Namespace).UpdateStatus(cluster)
	return err
}

func (a *Actuator) setMachinePhase(machine *clusterv1alpha1.Machine, phase string) (*clusterv1alpha1.Machine, error) {
	machine.Status.Phase = &phase
	return a.clusterClient.ClusterV1alpha1().Machines(machine.ObjectMeta.Namespace).UpdateStatus(machine)
}

func (a *Actuator) Delete(context context.Context, cluster *clusterv1alpha1.Cluster, machine *clusterv1alpha1.Machine) error {
	klog.Info("[proxmox] Deleting machine")

	machine, _ = a.setMachinePhase(machine, "Terminating")

	vmRef, err := a.getVMRef(machine)
	if err != nil {
		return err
	}
	if machineExists, err := a.Exists(context, cluster, machine); !machineExists && err == nil {
		return nil
	}
	if _, err = a.client.StopVm(vmRef); err != nil {
		return err
	}
	if _, err = a.client.DeleteVm(vmRef); err != nil {
		return err
	}
	clusterClient, err := a.getClusterConfig(cluster)
	if err != nil {
		return err
	}
	return clusterClient.CoreV1().Nodes().Delete(machine.ObjectMeta.Name, &v1.DeleteOptions{})
}

func (a *Actuator) Update(context.Context, *clusterv1alpha1.Cluster, *clusterv1alpha1.Machine) error {
	klog.Info("[proxmox] Updating machine")
	return nil
}

func (a *Actuator) Exists(context context.Context, cluster *clusterv1alpha1.Cluster, machine *clusterv1alpha1.Machine) (bool, error) {
	klog.Infof("[proxmox] Asking if machine %s exists", machine.ObjectMeta.Name)
	_, err := a.getVmInfo(machine)
	if err == nil {
		klog.Infof("[proxmox] Machine %q exists", machine.ObjectMeta.Name)
	} else {
		klog.Infof("[proxmox] Machine %q does not exist", machine.ObjectMeta.Name)
	}
	return err == nil, nil
}

func (a *Actuator) GetIP(cluster *clusterv1alpha1.Cluster, machine *clusterv1alpha1.Machine) (string, error) {
	klog.Infof("[proxmox] Asking for the IP address for machine %s", machine.ObjectMeta.Name)
	vmRef, err := a.getVMRef(machine)
	if err != nil {
		return "", err
	}
	networkInterfaces, err := a.client.GetVmAgentNetworkInterfaces(vmRef)
	if err != nil {
		return "", err
	}
	for _, networkInterface := range networkInterfaces {
		if networkInterface.Name == "lo" {
			continue
		}
		if len(networkInterface.IPAddresses) > 0 {
			ipAddr := networkInterface.IPAddresses[0].String()
			klog.Infof("[proxmox] Returning %q IP address for machine %s", ipAddr, machine.ObjectMeta.Name)
			return ipAddr, nil
		}
	}
	return "", fmt.Errorf("could not find a suitable IP address for machine %q in cluster %q", machine.ObjectMeta.Name, cluster.ObjectMeta.Name)
}

func (a *Actuator) GetKubeConfigContents(cluster *clusterv1alpha1.Cluster, machine *clusterv1alpha1.Machine) (string, error) {
	if !a.isControlPlane(machine) {
		return "", fmt.Errorf("machine %q is not a control plane; cannot retrieve administrative kubeconfig from this machine", machine.ObjectMeta.Name)
	}
	vmRef, err := a.getVMRef(machine)
	if err != nil {
		return "", err
	}
	kubeConfigContents, err := a.client.GetVmAgentFileRead(vmRef, "/etc/kubernetes/admin.conf")
	if err != nil {
		return "", err
	}
	return kubeConfigContents, nil
}

func (a *Actuator) getClusterConfig(cluster *clusterv1alpha1.Cluster) (kubernetes.Interface, error) {
	machineList, err := a.clusterClient.Cluster().Machines(cluster.ObjectMeta.Namespace).List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, machine := range machineList.Items {
		if machine.Labels[clusterv1alpha1.MachineClusterLabelName] == cluster.ObjectMeta.Name {
			if kubeConfigContents, err := a.GetKubeConfigContents(cluster, &machine); err == nil {
				if restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfigContents)); err == nil {
					if restClient, err := kubernetesclient.NewForConfig(restConfig); err == nil {
						return restClient, nil
					} else {
						klog.Warningf("could not intialize kubernetes client from REST config: %v", err)
					}
				} else {
					klog.Warningf("could not retrieve REST config from kubeconfig contents: %v", err)
				}
			} else {
				klog.Warningf("could not retrieve kubeconfig contents: %v", err)
			}
		}
	}
	return nil, errors.New("could not get cluster configuration")
}

func (a *Actuator) isControlPlane(machine *clusterv1alpha1.Machine) bool {
	return machine.Labels["set"] == "controlplane"
}

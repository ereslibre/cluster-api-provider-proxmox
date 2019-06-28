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

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen=github.com/ereslibre/cluster-api-provider-proxmox/pkg/apis/proxmoxproviderconfig
// +k8s:openapi-gen=true
// +k8s:defaulter-gen=TypeMeta
package v1alpha1

import (
	"encoding/json"
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/runtime/scheme"
	"sigs.k8s.io/yaml"
)

const GroupName = "proxmoxproviderconfig.k8s.io"

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "proxmoxproviderconfig.k8s.io", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// ClusterConfigFromProviderSpec unmarshals a provider config into an Proxmox Cluster type
func ClusterSpecFromProviderSpec(providerSpec clusterv1.ProviderSpec) (*ProxmoxClusterProviderSpec, error) {
	if providerSpec.Value == nil {
		return nil, errors.New("no such providerSpec found in manifest")
	}

	var config ProxmoxClusterProviderSpec
	if err := yaml.Unmarshal(providerSpec.Value.Raw, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// ClusterStatusFromProviderStatus unmarshals a provider status into an Proxmox Cluster Status type
func ClusterStatusFromProviderStatus(extension *runtime.RawExtension) (*ProxmoxClusterProviderStatus, error) {
	if extension == nil {
		return &ProxmoxClusterProviderStatus{}, nil
	}

	status := &ProxmoxClusterProviderStatus{}
	if err := yaml.Unmarshal(extension.Raw, status); err != nil {
		return nil, err
	}

	return status, nil
}

// This is the same as ClusterSpecFromProviderSpec but we
// expect there to be a specific Spec type for Machines soon
func MachineSpecFromProviderSpec(providerSpec clusterv1.ProviderSpec) (*ProxmoxMachineProviderSpec, error) {
	if providerSpec.Value == nil {
		return nil, errors.New("no such providerSpec found in manifest")
	}

	var config ProxmoxMachineProviderSpec
	if err := yaml.Unmarshal(providerSpec.Value.Raw, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func EncodeClusterStatus(status *ProxmoxClusterProviderStatus) (*runtime.RawExtension, error) {
	if status == nil {
		return &runtime.RawExtension{}, nil
	}

	var rawBytes []byte
	var err error

	//  TODO: use apimachinery conversion https://godoc.org/k8s.io/apimachinery/pkg/runtime#Convert_runtime_Object_To_runtime_RawExtension
	if rawBytes, err = json.Marshal(status); err != nil {
		return nil, err
	}

	return &runtime.RawExtension{
		Raw: rawBytes,
	}, nil
}

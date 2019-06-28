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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProxmoxMachineProviderSpec is the type that will be embedded in a Machine.Spec.ProviderSpec field
// for a Proxmox Instance. It is used by the Proxmox machine actuator to create a single machine instance.
// +k8s:openapi-gen=true
type ProxmoxMachineProviderSpec struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProxmoxClusterProviderSpec is the providerSpec for Proxmox in the cluster object
// +k8s:openapi-gen=true
type ProxmoxClusterProviderSpec struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProxmoxClusterProviderStatus contains the status fields
// relevant to Proxmox in the cluster object.
// +k8s:openapi-gen=true
type ProxmoxClusterProviderStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ProxmoxMachineProviderSpec{})
	SchemeBuilder.Register(&ProxmoxClusterProviderSpec{})
	SchemeBuilder.Register(&ProxmoxClusterProviderStatus{})
}

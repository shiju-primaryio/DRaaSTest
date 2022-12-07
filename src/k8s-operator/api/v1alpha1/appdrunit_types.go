/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AppDRUnitSpec defines the desired state of AppDRUnit
type AppDRUnitSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AppDRUnit. Edit appdrunit_types.go to remove/update
	Site              string   `json:"site,omitempty"`
	ProtectVMUUIDList []string `json:"protectvmuuidList,omitempty"`
	Description       string   `json:"description,omitempty"`
}

// AppDRUnitStatus defines the observed state of AppDRUnit
type AppDRUnitStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AppDRUnit is the Schema for the appdrunits API
type AppDRUnit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppDRUnitSpec   `json:"spec,omitempty"`
	Status AppDRUnitStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AppDRUnitList contains a list of AppDRUnit
type AppDRUnitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppDRUnit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppDRUnit{}, &AppDRUnitList{})
}

type VmPolicySchema struct {
	VmUuid     string `json:"vm_uuid,omitempty"`
	PolicyName string `json:"policy_name,omitempty"`
	VcHostIp   string `json:"vc_host_ip,omitempty"`
	VcUsername string `json:"vc_user_name,omitempty"`
	VcPassword string `json:"vc_password,omitempty"`
}

type IsPolicyExistsSchema struct {
	PolicyName string `json:"policy_name,omitempty"`
	VcHostIp   string `json:"vc_host_ip,omitempty"`
	VcUsername string `json:"vc_user_name,omitempty"`
	VcPassword string `json:"vc_password,omitempty"`
}

type PolicyDetails struct {
	IsPolicyExists bool   `json:"is_policy_exists,omitempty"`
	PolicyName     string `json:"policy_name,omitempty"`
	PolicyId       string `json:"policy_id,omitempty"`
}

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

	Site              string            `json:"site,omitempty"`
	RemoteSite        string            `json:"remoteSite,omitempty"`
	ProtectVMUUIDList []VmPolicyRequest `json:"protectvmuuidList,omitempty"`
	Description       string            `json:"description,omitempty"`
	VCenter           VCenterSpec       `json:"vCenter,omitempty"`
	// Application will run on RemoteSite when trigger failover is set to true.
	// TriggerFailover will invoke terraform script to create infra, get mapping of vmdks
	TriggerFailover bool   `json:"triggerFailover,omitempty"`
	VesToken        string `json:"vesToken,omitempty"`
}

/*
type VMDKFromPostGresDResponse struct {
	VMList []struct {
		VmdkId    string `json:"id"`
		VmdkScope string `json:"scope"`
	} `json:"data"`
}
*/

type VMDKListFromPostGresDResponse struct {
	Data []VMDKFromPostGresDResponse `json:"data,omitempty"`
}

type VMDKFromPostGresDResponse struct {
	VmdkId    string `json:"id"`
	VmdkScope string `json:"scope"`
}

/*
// VCenterSpec contains vCenter related connection info

	type VCenterSpec struct {
		IP string `json:"ip,omitempty"`
		// TODO: Change below fields to k8s secret
		UserName string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	}
*/
type TriggerFailoverVmdkMapping struct {
	VmName            string `json:"vmName,omitempty"`
	UnitNumber        int    `json:"unitNumber"`
	ScsiControllerId  string `json:"scsiControllerId,omitempty"`
	Label             string `json:"label,omitempty"`
	SourceVmdkID      string `json:"sourceVmdkID,omitempty"`
	TargetVmdkID      string `json:"targetVmdkID,omitempty"`
	SourceVmUUID      string `json:"sourceVmUUID,omitempty"`
	TargetVmUUID      string `json:"targetVmUUID,omitempty"`
	SourceScope       string `json:"sourceScope,omitempty"`
	TargetScope       string `json:"targetScope,omitempty"`
	SentBlocks        string `json:"sentBlocks,omitempty"`
	TotalBlocks       string `json:"totalBlocks,omitempty"`
	SentCT            string `json:"sentCT,omitempty"`
	Ack               string `json:"ack,omitempty"`
	ActiveFailover    bool   `json:"activeFailover"`
	FailoverTriggerID string `json:"failoverTriggerID"`
}

type VmPolicyRequest struct {
	VmUuid         string `json:"vmUuid,omitempty"`
	IsPolicyAttach bool   `json:"isPolicyAttach,omitempty"`
}

// AppDRUnitStatus defines the observed state of AppDRUnit
type AppDRUnitStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Site            string     `json:"site,omitempty"`
	ProtectedVmList []VMStatus `json:"protectedVmList,omitempty"`
	RemoteSite      string     `json:"remoteSite,omitempty"`

	FailoverStatus string `json:"failoverStatus,omitempty"`
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
	VmUuid     string `json:"vmUuid,omitempty"`
	PolicyName string `json:"policyName,omitempty"`
	VcHostIp   string `json:"vcHostIp,omitempty"`
	VcUsername string `json:"vcUserName,omitempty"`
	VcPassword string `json:"vcPassword,omitempty"`
}

type IsPolicyExistsSchema struct {
	PolicyName string `json:"policyName,omitempty"`
	VcHostIp   string `json:"vcHostIp,omitempty"`
	VcUsername string `json:"vcUserName,omitempty"`
	VcPassword string `json:"vcPassword,omitempty"`
}

type PolicyDetails struct {
	PolicyName string `json:"policyName,omitempty"`
	PolicyId   string `json:"policyId,omitempty"`
}

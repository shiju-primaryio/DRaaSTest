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
	// Application will run on RemoteSite when trigger failover is set to true.
	// TriggerFailover will invoke terraform script to create infra, get mapping of vmdks

	Site                               string            `json:"site,omitempty"`
	PeerSite                           string            `json:"peerSite,omitempty"`
	ProtectVMUUIDList                  []VmPolicyRequest `json:"protectvmuuidList,omitempty"`
	Description                        string            `json:"description,omitempty"`
	VCenter                            VCenterSpec       `json:"vCenter,omitempty"`
	TriggerFailover                    bool              `json:"triggerFailover,omitempty"`
	TriggerFailbackWithLiveReplication bool              `json:"triggerFailbackWithLiveReplication,omitempty"`
	TriggerCancelRecoveryOperation     bool              `json:"triggerCancelRecoveryOperation,omitempty"`
	TriggerFailback                    bool              `json:"triggerFailback,omitempty"`
	VesToken                           string            `json:"vesToken,omitempty"`
	//SnifPhpUrl                         string            `json:"snifPhpUrl,omitempty"`
}

type VMDKListFromPostGresDResponse struct {
	Data []VMDKFromPostGresDResponse `json:"data,omitempty"`
}

type VMDKFromPostGresDResponse struct {
	VmdkId         string `json:"id"`
	VmdkScope      string `json:"scope"`
	ReceivedIOs    string `json:"receivedIOs"`
	ReceivedBlocks string `json:"receivedBlocks"`
	TotalBlocks    string `json:"totalBlocks"`
}

/*
type VMDKActiveBitInfo struct {
	VmdkScope     string `json:"vmdkscope"`
	ActiveVMDKBit string `json:"activeVMDKBit"`
}

type VMActiveBitInfo struct {
	SourceVmUUID string              `json:"sourceVmUUID"`
	TargetVmUUID string              `json:"targetVmUUID"`
	ActiveVMBit  string              `json:"activeVMBit"`
	VmdkList     []VMDKActiveBitInfo `json:"vmdkList"`
}
*/
/*
// VCenterSpec contains vCenter related connection info

	type VCenterSpec struct {
		IP string `json:"ip,omitempty"`
		// TODO: Change below fields to k8s secret
		UserName string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	}
*/

type TriggerFailoverVmMapping struct {
	VmName          string                       `json:"vmName,omitempty"`
	SourceVmUUID    string                       `json:"sourceVmUUID,omitempty"`
	TargetVmUUID    string                       `json:"targetVmUUID,omitempty"`
	IsActiveBitTrue bool                         `json:"IsActiveBitTrue"`
	TriggerPowerOff bool                         `json:"triggerPowerOff"`
	VmdkStatusList  []TriggerFailoverVmdkMapping `json:"vmdkStatusList,omitempty"`
}

type TriggerFailoverVmdkMapping struct {
	UnitNumber        int    `json:"unitNumber"`
	ScsiControllerId  string `json:"scsiControllerId,omitempty"`
	Label             string `json:"label,omitempty"`
	SourceVmdkID      string `json:"sourceVmdkID,omitempty"`
	TargetVmdkID      string `json:"targetVmdkID,omitempty"`
	SourceScope       string `json:"sourceScope,omitempty"`
	TargetScope       string `json:"targetScope,omitempty"`
	SentBlocks        string `json:"sentBlocks,omitempty"`
	TotalBlocks       string `json:"totalBlocks,omitempty"`
	SentCT            string `json:"sentCT,omitempty"`
	Ack               string `json:"ack,omitempty"`
	ActiveFailover    bool   `json:"activeFailover"`
	FailoverTriggerID string `json:"failoverTriggerID"`
	RehydrationStatus string `json:"rehydrationStatus,omitempty"`
}

type TriggerFailbackVmMapping struct {
	VmName          string                       `json:"vmName,omitempty"`
	SourceVmUUID    string                       `json:"sourceVmUUID,omitempty"`
	TargetVmUUID    string                       `json:"targetVmUUID,omitempty"`
	IsActiveBitTrue bool                         `json:"IsActiveBitTrue"`
	TriggerPowerOff bool                         `json:"triggerPowerOff"`
	VmdkStatusList  []TriggerFailbackVmdkMapping `json:"vmdkStatusList,omitempty"`
}

type TriggerFailbackVmdkMapping struct {
	UnitNumber        int    `json:"unitNumber"`
	ScsiControllerId  string `json:"scsiControllerId,omitempty"`
	Label             string `json:"label,omitempty"`
	SourceVmdkID      string `json:"sourceVmdkID,omitempty"`
	TargetVmdkID      string `json:"targetVmdkID,omitempty"`
	SourceScope       string `json:"sourceScope,omitempty"`
	TargetScope       string `json:"targetScope,omitempty"`
	SentBlocks        string `json:"sentBlocks,omitempty"`
	TotalBlocks       string `json:"totalBlocks,omitempty"`
	SentCT            string `json:"sentCT,omitempty"`
	Ack               string `json:"ack,omitempty"`
	Follow_Seq        string `json:"follow_Seq,omitempty"`
	ActiveFailback    bool   `json:"activeFailback"`
	FailbackTriggerID string `json:"failbackTriggerID"`
	RehydrationStatus string `json:"rehydrationStatus,omitempty"`
}

type VmPolicyRequest struct {
	VmUuid         string `json:"vmUuid,omitempty"`
	IsPolicyAttach bool   `json:"isPolicyAttach,omitempty"`
}

// AppDRUnitStatus defines the observed state of AppDRUnit
type AppDRUnitStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Site                 string                     `json:"site,omitempty"`
	ProtectedVmList      []VMStatus                 `json:"protectedVmList,omitempty"`
	PeerSite             string                     `json:"peerSite,omitempty"`
	FailoverStatus       FailoverStatus             `json:"failoverStatus,omitempty"`
	FailbackStatus       FailbackStatus             `json:"failbackStatus,omitempty"`
	FailoverVmListStatus []TriggerFailoverVmMapping `json:"failoverVmListStatus,omitempty"`
	FailbackVmListStatus []TriggerFailbackVmMapping `json:"failbackVmListStatus,omitempty"`
}

/*
const (

	RECOVERY_ACTIVITY_NOT_STARTED   = "NOT_STARTED"
	RECOVERY_ACTIVITY_STARTED       = "STARTED"
	RECOVERY_ACTIVITY_IN_PROGRESS   = "IN_PROGRESS"
	RECOVERY_ACTIVITY_COMPLETED     = "COMPLETED"
	RECOVERY_ACTIVITY_ERROR_OCCURED = "ERROR_OCCURED"

)
*/

type InitiateFailbackRequest struct {
	SourceVMDKId string `json:"vmdk_id"`
	TargetVMDKId string `json:"new_vmdk_id"`
	Follow       bool   `json:"follow"`
}

type FailoverResponse struct {
	Data struct {
		Id           string `json:"id"`
		SourceVMDKId string `json:"vmdk_id"`
		TargetVMDKId string `json:"new_vmdk_id"`
		Ack          string `json:"ack"`
		Active       bool   `json:"active"`
		Sent_ct      string `json:"sent_ct"`
		Sentblocks   string `json:"sentblocks"`
		TotalBlocks  string `json:"totalBlocks"`
		Follow_Seq   string `json:"follow_Seq"`
	} `json:"data"`
}

type FailoverStatus struct {
	InfrastructureStatus  string `json:"infrastructureStatus"`
	PowerOffStatus        string `json:"powerOffStatus"`
	RehydrationStatus     string `json:"rehydrationStatus"`
	PowerOnStatus         string `json:"powerOnStatus"`
	OverallFailoverStatus string `json:"overallFailoverStatus"`
}

type FailbackStatus struct {
	InfrastructureStatus  string `json:"infrastructureStatus"`
	PowerOffStatus        string `json:"powerOffStatus"`
	RehydrationStatus     string `json:"rehydrationStatus"`
	PowerOnStatus         string `json:"powerOnStatus"`
	OverallFailbackStatus string `json:"overallFailbackStatus"`
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

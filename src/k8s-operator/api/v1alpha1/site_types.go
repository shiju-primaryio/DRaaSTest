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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SiteSpec defines the desired state of Site
type SiteSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	SiteAdmins    []rbacv1.Subject `json:"siteAdmins,omitempty"`
	IsPrimarySite bool             `json:"isPrimarySite,omitempty"`
	PeerSite      string           `json:"peerSite,omitempty"`
	VCenter       VCenterSpec      `json:"vCenter,omitempty"`
	VMList        []VMSpec         `json:"vmList,omitempty"`
	// StoragePolicy should be deleted on deletion of Site
	StoragePolicy StoragePolicySpec `json:"storagePolicySpec,omitempty"`
}

// VCenterSpec contains vCenter related connection info
type VCenterSpec struct {
	IP string `json:"ip,omitempty"`
	// TODO: Change below fields to k8s secret
	UserName string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// VMSpec contains VM specifications
type VMSpec struct {
	UUID      string `json:"uuid,omitempty"`
	IsPowerOn bool   `json:"isPowerOn,omitempty"`
}

// StoragePolicySpec contains Storage Policy specs
type StoragePolicySpec struct {
	Name         string `json:"policy_name,omitempty"`
	Description  string `json:"policy_description,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
	Namespace    string `json:"namespace,omitempty"`
	NamespaceId  string `json:"namespace_id,omitempty"`
	Blocksize    int32  `json:"blocksize,omitempty"`
	Queuesize    int32  `json:"queuesize,omitempty"`
	Outsize      int32  `json:"outsize,omitempty"`
	Port         string `json:"port,omitempty"`
	User         string `json:"user,omitempty"`
	Secret       string `json:"secret,omitempty"`
	Host         string `json:"host,omitempty"`
	Enckey       string `json:"enckey,omitempty"`
}

// SiteStatus defines the observed state of Site
type SiteStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	SiteState SiteState  `json:"state,omitempty"`
	VmList    []VMStatus `json:"vmList,omitempty"`
	PolicyId  string     `json:"policyId,omitempty"`
	TempId    int        `json:"tmpId,omitempty"`
	Error     ErrorField `json:"error,omitempty"`
}

type SiteState string

const (
	SiteStateUp   SiteState = "Up"
	SiteStateDown SiteState = "Down"
)

// VMStatus contains VM's current status
type VMStatus struct {
	VmId        string   `json:"vm_id,omitempty"`
	Name        string   `json:"name,omitempty"`
	VmUuid      string   `json:"vm_uuid"`
	CPUs        int32    `json:"cpus,omitempty"`
	MemoryMB    int32    `json:"memory_mb,omitempty"`
	GuestID     string   `json:"guest_id,omitempty"`
	Disks       []Disk   `json:"disks,omitempty"`
	IsProtected bool     `json:"is_protected"`
	IpAddress   []string `json:"ip_address,omitempty"`
	NumDisks    int      `json:"num_disks,omitempty"`
	IsPowerOn   bool     `json:"isPowerOn"`
	PowerState  string   `json:"power_state,omitempty"`
}

// Disk configuration
type Disk struct {
	Name            string `json:"file_name"`
	VmId            string `json:"vm_id,omitempty"`
	Datastore       string `json:"datastore,omitempty"`
	ThinProvisioned bool   `json:"thin_provisioned"`
	SizeMB          int64  `json:"size_mb,omitempty"`
	UnitNumber      int32  `json:"unit_number"`
	Label           string `json:"label,omitempty"`
	IofilterName    string `json:"iofilter_name"`
}

type ErrorField struct {
	ErrorMessage string `json:"errorMsg,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Site is the Schema for the sites API
type Site struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SiteSpec   `json:"spec,omitempty"`
	Status SiteStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SiteList contains a list of Site
type SiteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Site `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Site{}, &SiteList{})
}

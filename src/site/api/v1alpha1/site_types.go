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

	SiteAdmins []rbacv1.Subject `json:"siteAdmins,omitempty"`
	VCenter    VCenterSpec      `json:"vCenter,omitempty"`
	VMList     []VMSpec         `json:"vmList,omitempty"`
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
	Name string `json:"name,omitempty"`
	// TODO: ProtectVM List should be handled by DRUnit controller
	ProtectVM bool `json:"protect,omitempty"`
}

// SiteStatus defines the observed state of Site
type SiteStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	SiteState SiteState  `json:"state,omitempty"`
	VMList    []VMStatus `json:"vmList,omitempty"`
}

type SiteState string

const (
	SiteStateUp   SiteState = "Up"
	SiteStateDown SiteState = "Down"
)

// VMStatus contains VM's current status
type VMStatus struct {
	Name        string `json:"name,omitempty"`
	CPUs        int    `json:"cpus,omitempty"`
	Memory      int    `json:"memory,omitempty"`
	GuestID     string `json:"guestID,omitempty"`
	Disks       []Disk `json:"disks,omitempty"`
	IsProtected bool   `json:"isProtected,omitempty"`
}

// Disk configuration
type Disk struct {
	UnitNumber      int    `json:"unitNumber,omitempty"`
	Size            int    `json:"size,omitempty"`
	Label           string `json:"label,omitempty"`
	ThinProvisioned bool   `json:"thinProvisioned,omitempty"`
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

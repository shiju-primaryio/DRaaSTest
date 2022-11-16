
variable vpcmod_ibmcloudapikey {
  description = "The IBM Cloud platform API key needed to deploy IAM enabled resources"
  type        = string
   
}

variable ibm_region {
    description = "IBM Cloud region where all resources will be deployed"
    type        = string
    #default = "eu-de"
    #default = "us-east"
    default = "us-south"
}

variable ibm_zone_esxi_image {
    description = "IBM Cloud zone region to be used to use esxi image."
    type        = string
    #default = "us-east-2"
    #default = "eu-de-2"
    default = "us-south-3"
}

variable resource_group {
    description = "Name of resource group to create VPC"
    type        = string
    #default     = "terraform-rg"
     default     = "5f28b5cec9b143a1b58e7bd1ea96073a"
}
###################################

variable "unique_id" {
    default = "alpha-demo-tf"
  
}

variable generation {
  description = "generation for VPC. Can be 1 or 2"
  type        = number
  default     = 2
}


variable "ssh_key" {
    #default = "ubuntu-jumphost"
    #default = "wdc-bastion-key-same-as-dallas"
    default = "dev-bastionhostkey"
}

variable "jumphost_profile" {
  default = "cx2-2x4"
  }



variable classic_access {
  description = "Enable VPC Classic Access. Note: only one VPC per region can have classic access"
  type        = bool
  default     = false
}


variable enable_public_gateway {
  description = "Enable public gateways for subnets, true or false"
  type        = bool
  default     = true
}

variable security_group_rules {
  description = "List of security group rules for default VPC security group"
  default     = [
    {
      source    = "0.0.0.0/0"
      direction = "inbound"
    }
  ]
}

variable zone_1_cidr_blocks {
  description = "A list of zone 1 subnet IPs"
  default = "172.16.10.0/24" 

}
 variable zone_1_acl_rules {
  description = "Access control list rule set for zone 1 subnets"
  default = [
    {
      name        = "allow-all-inbound"
      action      = "allow"
      source      = "0.0.0.0/0"
      destination = "0.0.0.0/0"
      direction   = "inbound"
    },
    {
      name        = "allow-all-outbound"
      action      = "allow"
      source      = "0.0.0.0/0"
      destination = "0.0.0.0/0"
      direction   = "outbound"
    }
  ]
}


variable zone_2_cidr_blocks {
  description = "A list of zone 2 subnet IPs"
  default = "172.16.11.0/24"
 
}

variable zone_2_acl_rules {
  description = "Access control list rule set for zone 2 subnets"
  default = [
    {
      name        = "allow-all-inbound"
      action      = "allow"
      source      = "0.0.0.0/0"
      destination = "0.0.0.0/0"
      direction   = "inbound"
    },
    {
      name        = "allow-all-outbound"
      action      = "allow"
      source      = "0.0.0.0/0"
      destination = "0.0.0.0/0"
      direction   = "outbound"
    }
  ]
}

# Vcenter gateway should be changed
variable zone_3_cidr_blocks {
  description = "A list of zone 3 subnet IPs"
  default = "172.16.12.0/24"
 
}

############################
variable "windows_image" {
    #default = "ibm-windows-server-2022-full-standard-amd64-4"
   default = "ibm-windows-server-2019-full-standard-amd64-10"
  
}

data "ibm_is_image" "os" {
  name = var.windows_image
}

variable "linux_image" {    
   default = "ibm-ubuntu-22-04-1-minimal-amd64-1"
  
}

data "ibm_is_image" "linux" {
  name = var.linux_image
}

#############################
#######ESXI variables########

variable "vmw_host_profile" {
  default = "bx2d-metal-96x384"
  #default = "bx2-metal-96x384"
  #default = "cx2-metal-96x192"
  
}

variable "esxi_image" {
  description = "Base ESXI image name, terraform will find the latest available image id."
  default = "esxi-7-byol"
  type = string
}

variable "esxi_image_name" {
  description = "Use a specific ESXI image version to use for the hosts to override the latest by name."
  default = "ibm-esxi-7-0u3d-19482537-byol-amd64-1" 
  type = string
}

variable "host_vlan_id" {
  description = "VLAN ID for host network"
  default     = 0 
  type = number
}

variable "mgmt_vlan_id" {
  description = "VLAN ID for management network"
  default     = 100
  type = number
}

variable "mgmt_vlan_name" {
   default = "vlan-nic-vcenter"
   type = string
  
}

variable "vcenter_password" {
  description = "Define a common password for all elements. Optional, leave empty to get random passwords."
  default = ""
  type = string
}


variable "esxi_password" {
  description = "Define a common password for all elements. Optional, leave empty to get random passwords."
  default = ""
  type = string
}

variable "vmw_host_list" {
  default = 1
  
}

variable "esxi-hostname" {
  
  default = "172.16.11.5"

}

variable "esxi-username" {
  
  default = "root"
}

variable "esxi-password" {
  default = "PrimaryIO@123"
}

variable "esxi-deployment-nw" {
  default = "pg-mgmt"

}

variable "esxi-datastore" {
  default = "datastore1"  
}

variable "vcenter-ip" { 
  default = "172.16.11.7"
}

variable "vcenter-subnet-prefix" {
  default = "24"
}

variable "vcenter-gw" {
  default = "172.16.12.1"
}

variable "vcenter-password" {
  default = "PrimaryIO@123"
}

variable "vcenter-domain_name" {
  default = "primaryio.cloud"
}

variable "esxi-hostname-fqdn" { 
  default = "esxi.primaryio.cloud"
  
}

variable "vm-network1-vlan-id" {
  description = "VLAN ID for vm network"
  default     = 101
  type = number
}

variable "vm-network1-vlan-name" {
   default = "vlan-nic-vm1"
   type = string
}

variable "dallas-tgw-name" {
   default = "vpcdev-vmw-transitGW-LR"
   type = string
}

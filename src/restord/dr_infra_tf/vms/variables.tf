variable "vmsmod_vsphere_server" { }
variable "vsphere_server" {
    type        = string
    default = ""
}
variable "vsphere_user" {
    type        = string
    default = "administrator@primaryio.cloud"
}
variable "vsphere_password" {
    type        = string
    default = "PrimaryIO@123"
}
variable "datacentername" {
    type        = string
    default = "Datacenter1"
}
variable "datastorename" {
    type        = string
    default = "datastore1"
}
variable "clustername" {
    type        = string
    default = "Cluster1"
}
variable "resourcepoolname" {
    type        = string
    default = "ProtectIO_POOL"
}
variable "esxihostname" {
    type        = string
    default = ""
}
variable "esxipassword" {
    type        = string
    default = "PrimaryIO@123"
}
variable "vsphere_networkname" {
    type        = string
    default = "VM Network"
}

variable vmlist {
  description = "Map of VMs for configuration."
  type = list(object({
    name  = string
    num_cpus = number
    memory   = number
    guest_id = string
    disks    = list(object({
	unit_number = number
	size = number
	label = string
	thin_provisioned = bool
	}))
  }))
}



variable ibmcloud_api_key {}
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


## Build VM
resource "vsphere_datacenter" "dc" {
  name = "${var.datacentername}"
}

data "vsphere_datacenter" "dc" {
  name = "${var.datacentername}"
  depends_on = [vsphere_datacenter.dc]
}

resource "vsphere_compute_cluster" "cluster" {
  name = "${var.clustername}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
  drs_enabled          = true
  drs_automation_level = "fullyAutomated"
  ha_enabled = false
}

data "vsphere_compute_cluster" "cluster" {
  name = "${var.clustername}"
  datacenter_id = data.vsphere_datacenter.dc.id
  depends_on = [vsphere_compute_cluster.cluster]
}

#get latest thumbprint of ESXi host
data "vsphere_host_thumbprint" "thumbprint-esxi" {
  address = "${var.esxihostname}"
  insecure = true
}


#Add ESXi host to the cluster
resource vsphere_host "host" {
  hostname = "${var.esxihostname}"
  username = "root"
  password = "${var.esxipassword}"
  cluster = data.vsphere_compute_cluster.cluster.id
  depends_on = [vsphere_compute_cluster.cluster]
  thumbprint = data.vsphere_host_thumbprint.thumbprint-esxi.id
}

data "vsphere_host" "host" {
  name = "${var.esxihostname}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
  depends_on = [vsphere_host.host]
}

data "vsphere_datastore" "datastore" {
  name = "${var.datastorename}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
  depends_on = [vsphere_host.host]
}

resource "vsphere_resource_pool" "pool" {
  name = "${var.resourcepoolname}"
  parent_resource_pool_id = data.vsphere_compute_cluster.cluster.resource_pool_id
}

data "vsphere_resource_pool" "pool" {
  name = "${var.resourcepoolname}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
  depends_on = [vsphere_resource_pool.pool]
}

data "vsphere_network" "mgmt_lan" {
  name = "${var.vsphere_networkname}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
  depends_on = [vsphere_host.host]
}

resource "vsphere_virtual_machine" "test2" {
  for_each = {for i, v in var.vmlist:  i => v}
  name             = each.value.name
  resource_pool_id = "${data.vsphere_resource_pool.pool.id}"
  datastore_id     = "${data.vsphere_datastore.datastore.id}"
  host_system_id   = "${data.vsphere_host.host.id}"

  num_cpus   = each.value.num_cpus
  memory     = each.value.memory
  wait_for_guest_net_timeout = 0

  guest_id = each.value.guest_id
  nested_hv_enabled =true
  network_interface {
   network_id     = "${data.vsphere_network.mgmt_lan.id}"
   adapter_type   = "vmxnet3"
  }
 
  dynamic "disk" {
    for_each = each.value.disks
    content {
      unit_number      = disk.value.unit_number
      size = disk.value.size
      label = disk.value.label
      eagerly_scrub    = false
      thin_provisioned = disk.value["thin_provisioned"]
    }
  }

}
resource "local_file" "vcenter_details" {
    content  = var.vmsmod_vsphere_server
    filename = "vcenter_details.txt"
}


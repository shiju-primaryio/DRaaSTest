terraform {
  required_providers {
    ibm = {
      source = "IBM-Cloud/ibm"
      #version = "1.19.0"
      version = "1.46.0"

    }
  }
}

/*
module "vpc_esxi_vcenter" {
  source = "./vpc_esxi_vcenter"
  #ibmcloud_api_key =var.ibmcloud_api_key
  vpcmod_ibmcloudapikey =var.ibmcloud_api_key
}
*/
locals {
 //vmlist = jsondecode(file("${path.module}/CreateModifyVmList.json"))["vmlist"]
 vmlist = jsondecode(file("/home/ubuntu/DRaaS/src/restord/dr_infra_vm/CreateModifyVmList.json"))["vmlist"]
}

#data ibm_is_bare_metal_server_network_interface {

#}
module "vms" {
  source = "./vms"
  vmsmod_vsphere_server = "172.16.12.5"
  #vmsmod_vsphere_server = module.vpc_esxi_vcenter.vcenter-privateip
  #vsphereserver = module.vpc_esxi_vcenter.vcenter-privateip
  vmlist=local.vmlist
  #vmlist="${var.vmlist}"
  esxihostname = "172.16.12.4"
  #esxihostname = module.vpc_esxi_vcenter.ESXi-privateip
  #esxi_privateip = module.vpc_esxi_vcenter.ESXi-privateip


#  vsphere_server = module.vpc_esxi_vcenter.ibm_is_bare_metal_server_network_interface.esx_host_vcenter.primary_ip[0].address
  #ESXi-privateip = module.vpc_esxi_vcenter.ESXi-privateip
#ibm_is_bare_metal_server_network_interface.esx_host_vcenter.primary_ip[0].address
  #ibmcloud_api_key ="${var.ibmcloud_api_key}"
  #vpc_id = module.vpc_esxi_vcenter.vpc_id
  #subnet_ids = [ module.vpc.subnet_a_id , module.vpc.subnet_b_id, module.vpc.subnet_c_id ]
}

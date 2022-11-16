
resource "ibm_is_vpc" "vpc-test" {
    name = "${var.unique_id}-vpc"
    address_prefix_management = "manual"
    resource_group = var.resource_group
    
    }


resource "ibm_is_vpc_address_prefix" "subnet_prefix1" {
        cidr        = var.zone_1_cidr_blocks       
        name        = "${var.unique_id}-prefix-zone-1"
        vpc         =  ibm_is_vpc.vpc-test.id
        zone        = "${var.ibm_region}-1"
     } 

     

resource "ibm_is_subnet" "subnet1" {
    name                     = "${var.unique_id}-subnet1"
    vpc                      = ibm_is_vpc.vpc-test.id
    zone                     = "${var.ibm_region}-1"
    ipv4_cidr_block          = ibm_is_vpc_address_prefix.subnet_prefix1.cidr
    public_gateway           = join("", ibm_is_public_gateway.gateway1.*.id)
}


resource "ibm_is_vpc_address_prefix" "subnet_prefix2" {
        cidr        = var.zone_2_cidr_blocks       
        name        = "${var.unique_id}-prefix-zone-2"
        vpc         =  ibm_is_vpc.vpc-test.id
        zone        = "${var.ibm_region}-2"
     } 
  
resource "ibm_is_subnet" "subnet2" {
    name                     = "${var.unique_id}-subnet2"
    vpc                      = ibm_is_vpc.vpc-test.id
    zone                     = "${var.ibm_region}-2" 
    ipv4_cidr_block          = ibm_is_vpc_address_prefix.subnet_prefix2.cidr
    public_gateway           = ibm_is_public_gateway.gateway2.id
}

resource "ibm_is_vpc_address_prefix" "subnet_prefix3" {
        cidr        = var.zone_3_cidr_blocks       
        name        = "${var.unique_id}-prefix-zone-3"
        vpc         =  ibm_is_vpc.vpc-test.id
        zone        = "${var.ibm_region}-3"
     } 
  
resource "ibm_is_subnet" "subnet3" {
    name                     = "${var.unique_id}-subnet3"
    vpc                      = ibm_is_vpc.vpc-test.id
    zone                     = "${var.ibm_region}-3" 
    ipv4_cidr_block          = ibm_is_vpc_address_prefix.subnet_prefix3.cidr
    public_gateway           = ibm_is_public_gateway.gateway3.id
}


######################################
#publicGW

resource "ibm_is_public_gateway" "gateway1" { 
        name    = "${var.unique_id}-pgw1"
        resource_group = var.resource_group
        vpc = ibm_is_vpc.vpc-test.id
        zone = "${var.ibm_region}-1" 
}


resource "ibm_is_public_gateway" "gateway2" { 
        name    = "${var.unique_id}-pgw2"
        resource_group = var.resource_group
        vpc = ibm_is_vpc.vpc-test.id
        zone = "${var.ibm_region}-2" 
}

resource "ibm_is_public_gateway" "gateway3" { 
        name    = "${var.unique_id}-pgw3"
        resource_group = var.resource_group
        vpc = ibm_is_vpc.vpc-test.id
        zone = "${var.ibm_region}-3" 
}

###################

resource "ibm_is_security_group" "jumphost-sg" {
    name = "${var.unique_id}-jumphost-sg"
    vpc  = ibm_is_vpc.vpc-test.id
}

# allow all incoming network traffic on port 22
resource "ibm_is_security_group_rule" "ingress_rdp_all" {
    group     = ibm_is_security_group.jumphost-sg.id
    direction = "inbound"
    remote    = "0.0.0.0/0"

    tcp {
      port_min = 3389
      port_max = 3389
    }
}

resource "ibm_is_security_group_rule" "ingress_ssh_all" {
    group     = ibm_is_security_group.jumphost-sg.id
    direction = "inbound"
    remote    = "52.118.189.118/32"

    tcp {
      port_min = 22
      port_max = 22
    }
}

resource "ibm_is_security_group_rule" "ingress_internal_all" {
    group     = ibm_is_security_group.jumphost-sg.id
    direction = "inbound"
    remote    = "10.0.0.0/8"

    tcp {
      port_min = 1
      port_max = 65535
    }
}
resource "ibm_is_security_group_rule" "ingress_vpc1_all" {
    group     = ibm_is_security_group.jumphost-sg.id
    direction = "inbound"
    remote    = var.zone_1_cidr_blocks

    tcp {
      port_min = 1
      port_max = 65535
    }
}
resource "ibm_is_security_group_rule" "ingress_vpc2_all" {
    group     = ibm_is_security_group.jumphost-sg.id
    direction = "inbound"
    remote    = var.zone_2_cidr_blocks

    tcp {
      port_min = 1
      port_max = 65535
    }
}

resource "ibm_is_security_group_rule" "ingress_vpc3_all" {
    group     = ibm_is_security_group.jumphost-sg.id
    direction = "inbound"
    remote    = var.zone_3_cidr_blocks

    tcp {
      port_min = 1
      port_max = 65535
    }
}




resource "ibm_is_security_group_rule" "egress_rdp_all" {
    group     = ibm_is_security_group.jumphost-sg.id
    direction = "outbound"
    remote    = "0.0.0.0/0"

    tcp {
      port_min = 1
      port_max = 65535
    }
}

resource "ibm_is_security_group_rule" "egress_udp_all" {
    group     = ibm_is_security_group.jumphost-sg.id
    direction = "outbound"
    remote    = "0.0.0.0/0"

    udp {
      port_min = 1
      port_max = 65535
    }
}

data "ibm_is_ssh_key" "ssh_key_id" {
    name  = var.ssh_key
}

resource "ibm_is_instance" "linux-jumphost" {
    name             = "${var.unique_id}-linux-jumphost"
    vpc              = ibm_is_vpc.vpc-test.id
    zone             = "${var.ibm_region}-1"
    resource_group   = var.resource_group
    keys             = [data.ibm_is_ssh_key.ssh_key_id.id]    
    image            = data.ibm_is_image.linux.id
    #user_data       = file("./install-vcenter.tftpl")
    user_data        = templatefile("vpc_esxi_vcenter/install-vcenter.tftpl", {
                     ibmcloud_api_key = var.vpcmod_ibmcloudapikey
                     esxi-hostname =  ibm_is_bare_metal_server.esx_host.primary_network_interface[0].primary_ip[0].address
                     esxi-username = var.esxi-username
                     esxi-password = var.esxi-password 
                     esxi-deployment-nw = var.esxi-deployment-nw
                     esxi-datastore = var.esxi-datastore
                     vcenter-ip  = ibm_is_bare_metal_server_network_interface.esx_host_vcenter.primary_ip[0].address
                     vcenter-subnet-prefix = var.vcenter-subnet-prefix
                     vcenter-gw   = var.vcenter-gw
                     vcenter-password  = var.vcenter-password
                     vcenter-domain_name   =  var.vcenter-domain_name
                     
    })
    profile          = var.jumphost_profile
    depends_on = [ibm_is_bare_metal_server.esx_host]

    primary_network_interface {
        subnet          = ibm_is_subnet.subnet1.id
        security_groups = [ibm_is_security_group.jumphost-sg.id]
    }

}


resource "ibm_is_floating_ip" "jumphost_fip" {
    name   = "${var.unique_id}-jumphost-fip"
    target = ibm_is_instance.linux-jumphost.primary_network_interface[0].id
   
    }


##############################################################
# Calculate the most recently available OS Image Name for the
# OS Provided
##############################################################



data "ibm_is_images"  "os_images" {
    visibility = "public"
}

locals {
    os_images_filtered_esxi = [
        for image in data.ibm_is_images.os_images.images:
            image if ((image.os == var.esxi_image) && (image.status == "available"))
    ]
}

data "ibm_is_image" "vmw_esx_image" {
  name = var.esxi_image_name == "" ? local.os_images_filtered_esxi[0].name : var.esxi_image_name
}


resource "ibm_is_bare_metal_server" "esx_host" {
    profile         = var.vmw_host_profile
    user_data        = templatefile("vpc_esxi_vcenter/setup-esxi.tftpl", {                     
                     esxi-hostname-fqdn = var.esxi-hostname-fqdn
                     mgmt_vlan_id = var.mgmt_vlan_id
                     vm-network1-vlan-id  = var.vm-network1-vlan-id 
                     esxi-password  = var.esxi-password
                     esxi-deployment-nw  = var.esxi-deployment-nw
                     #ibmcloud_api_key = var.ibmcloud_api_key
                     ibmcloud_api_key = var.vpcmod_ibmcloudapikey
    })    
    name            = "${var.unique_id}-esxi01"
    resource_group  = var.resource_group
    image            = data.ibm_is_image.vmw_esx_image.id 
    zone            = "${var.ibm_region}-3"
    keys            = [data.ibm_is_ssh_key.ssh_key_id.id]
    
    primary_network_interface {
      # pci 1
      subnet        = ibm_is_subnet.subnet3.id
      allowed_vlans = ["100", "101"]
      name          = "pci-vmnic0-vmk0"
      security_groups = [ibm_is_security_group.jumphost-sg.id]
      enable_infrastructure_nat = true
    }   

    #tags = var.vmw_tags

    vpc =  ibm_is_vpc.vpc-test.id
    timeouts {
      create = "60m"
      update = "60m"
      delete = "60m"
    }

    lifecycle {
      ignore_changes = [user_data,image]
    }

}

resource "ibm_is_bare_metal_server_network_interface" "esx_host_vcenter" {    
    bare_metal_server = ibm_is_bare_metal_server.esx_host.id
    subnet = ibm_is_subnet.subnet3.id
    name   = var.mgmt_vlan_name
    security_groups = [ibm_is_security_group.jumphost-sg.id]
    allow_ip_spoofing = false
    vlan = var.mgmt_vlan_id
    allow_interface_to_float = true    
    depends_on = [ibm_is_bare_metal_server.esx_host]
}


resource "ibm_is_bare_metal_server_network_interface" "esx_host_vm1" {    
    bare_metal_server = ibm_is_bare_metal_server.esx_host.id
    subnet = ibm_is_subnet.subnet3.id
    name   = var.vm-network1-vlan-name
    security_groups = [ibm_is_security_group.jumphost-sg.id]
    allow_ip_spoofing = false
    vlan = var.vm-network1-vlan-id
    allow_interface_to_float = true    
    depends_on = [ibm_is_bare_metal_server.esx_host]
}

data "ibm_tg_gateway" "dallas-local-tgw" {
  name = var.dallas-tgw-name
}

resource "ibm_tg_connection" "dallas_local" {
  gateway     = data.ibm_tg_gateway.dallas-local-tgw.id
  network_type = "vpc"
  name        = "${var.unique_id}-connection"
  network_id  = ibm_is_vpc.vpc-test.resource_crn
  depends_on = [ibm_is_vpc.vpc-test]
}

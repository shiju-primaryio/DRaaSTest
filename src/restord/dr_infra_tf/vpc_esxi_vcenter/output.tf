
# output "esxi_private_ip"{

#   value = ibm_is_bare_metal_server.esx_host.primary_network_interface[0].primary_ip[0].address
# }

##############################################################
# Get host root passwords
##############################################################
data "local_file" "ssh_private_key_file" {
    filename = "${path.module}/private-key.pem"
}

#data "ibm_is_bare_metal_server_initialization" "esx_host01_init_values" {     
#    bare_metal_server = ibm_is_bare_metal_server.esx_host.id
#    private_key = data.local_file.ssh_private_key_file.content
#}


# output "ibm_is_bare_metal_server_initialization" {
#   description = "Access information for ESXi."
#   value = data.ibm_is_bare_metal_server_initialization.esx_host01_init_values
#   sensitive = true
# }

 
###############

#  data "ibm_is_instance_initialization" "win_jump_init_values" {     
#      bare_metal_server = ibm_is_instance.win-jumphost.id  
#      private_key = data.local_file.ssh_private_key_file.content
#  }



# output "ibm_is_instance_initialization"  {

#   value = data.ibm_is_instance_initialization.win_jump_init_values
# }


#   output "sshcommand" {
#     value = "ssh root@${ibm_is_floating_ip.fip1.address}"
#     }

output "vcenter-privateip" {
    value = ibm_is_bare_metal_server_network_interface.esx_host_vcenter.primary_ip[0].address
}

output "ESXi-privateip" {
    value = ibm_is_bare_metal_server.esx_host.primary_network_interface[0].primary_ip[0].address
}

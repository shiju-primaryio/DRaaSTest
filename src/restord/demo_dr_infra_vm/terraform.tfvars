
ibmcloud_api_key = "snFdJAls4oHKXKB5Adx9PnwiGyfgHM1_hCp_xpEMQpPP"
vmlist =[{
    name  = "test-vm5",
    num_cpus = 1,
    memory   = 2048,
    guest_id = "centos7_64Guest",
    disks    = [
	{
	unit_number=0
	size = 16,
	label = "test-vm5disk1.vmdk",
        policy_name ="abcd",
	thin_provisioned = true,
	},
	{
	unit_number=1
	size = 32,
        policy_name ="abcd",
	label = "test-vm5disk2.vmdk",
	thin_provisioned = true,
	}
	]
    },
    {
    name  = "test-vm6",
    num_cpus = 2,
    memory   = 4096,
    guest_id = "centos7_64Guest",
    disks    = [
	{
	unit_number=0
	size = 64,
        policy_name ="abcd",
	label = "test-vm6disk1.vmdk",
	thin_provisioned = true,
	},
	{
	unit_number=1
	size = 8,
        policy_name ="abcd",
	label = "test-vm6disk2.vmdk",
	thin_provisioned = true,
	}
	]
    }
]
#}

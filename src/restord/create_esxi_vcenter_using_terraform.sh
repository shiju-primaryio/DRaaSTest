#!/bin/bash

#mkdir  ${site}-vpc-esxi-vcenter-vms
#cp -rvd vpc-esxi-vcenter-vms  ${tenant}-vpc-esxi-vcenter-vms 
#cd ${tenant}-vpc-esxi-vcenter-vms

cd dr_infra_tf
date
terraform init 
terraform plan -target module.vpc_esxi_vcenter  
terraform apply -target module.vpc_esxi_vcenter  -auto-approve
date

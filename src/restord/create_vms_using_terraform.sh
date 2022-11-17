#!/bin/bash

cd dr_infra_tf
date
pwd= `pwd`
if [ ! -f .terraform/modules/modules.json ]
then
    terraform init
fi
terraform plan -target module.vms 
terraform apply -target module.vms -auto-approve
#terraform plan -target module.vms -chdir "$pwd" 
#terraform apply -target module.vms  -auto-approve -chdir "$pwd" 
date

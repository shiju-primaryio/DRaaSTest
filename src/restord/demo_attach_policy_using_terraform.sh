#!/bin/bash

terra_logs_dir="/home/ubuntu/terraform_logs"
mkdir -p "$terra_logs_dir" 
today_date_time=`date '+%Y_%m_%d__%H_%M_%S'`;
filename=$terra_logs_dir/policy_attach_terraform_output_"$today_date_time.txt"
echo $filename;

cd demo_dr_infra_vm
date|tee -a $filename 
pwd= `pwd`
if [ ! -f .terraform/modules/modules.json ]
then
    terraform init
fi
terraform plan -target module.vms |tee -a $filename 
terraform apply -target module.vms -auto-approve |tee -a $filename 
#terraform plan -target module.vms -chdir "$pwd" 
#terraform apply -target module.vms  -auto-approve -chdir "$pwd" 
date|tee -a $filename 

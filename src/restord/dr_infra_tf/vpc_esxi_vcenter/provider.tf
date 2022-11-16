terraform {
  required_providers {
    ibm = {
      source = "IBM-Cloud/ibm"
      #version = "1.19.0"
      version = "1.46.0"
      
    }
  }
}

provider "ibm" {
  region = "${var.ibm_region}"
  ibmcloud_api_key = var.vpcmod_ibmcloudapikey
}


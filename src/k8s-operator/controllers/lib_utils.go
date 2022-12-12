package controllers

import (
	"context"
	"fmt"
	"net/url"

	draasv1alpha1 "github.com/CacheboxInc/DRaaS/src/k8s-operator/api/v1alpha1"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/vim25"
)

// Obtain a client to call vCenter APIs
func GetVCenterClient(vcenter draasv1alpha1.VCenterSpec) (*vim25.Client, error) {
	ctx := context.Background()
	urlString := "https://" + vcenter.UserName + ":" + vcenter.Password + "@" + vcenter.IP + "/sdk"
	vCenterUrl, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	insecure := true
	/* if os.Getenv("VCENTER_INSECURE") != "1" {
		insecure = false
	} */
	session := cache.Session{URL: vCenterUrl, Insecure: insecure}
	client := new(vim25.Client)
	err = session.Login(ctx, client, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Logged in to vCenter.")
	return client, nil
}

func GetVmObject(vCenterClient *vim25.Client, vmUuid string) (object.VirtualMachine, error) {
	var vmObj object.VirtualMachine
	ctx := context.Background()

	f := find.NewFinder(vCenterClient, true)
	// Find one and only datacenter
	dc, err := f.DefaultDatacenter(ctx)
	if err != nil {
		fmt.Println("Error getting datacenter : ", err)
		return vmObj, err
	}

	// Make future calls local to this datacenter
	f.SetDatacenter(dc)
	vms, err := f.VirtualMachineList(ctx, "*")
	if err != nil {
		fmt.Println("Error getting VM list : ", err)
		return vmObj, err
	}

	for _, vm := range vms {
		if vm.UUID(ctx) == vmUuid {
			vmObj = *vm
		}
	}

	return vmObj, nil
}

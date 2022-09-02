package vm

import (
	// "github.com/vmware/govmomi/examples"
	// "github.com/vmware/govmomi/view"
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
)

func PrintVmInfo(vmName string) {
	ctx := context.Background()
	vCenterClient, err := GetVCenterClient(ctx)
	if err != nil {
		return
	}
	manager := view.NewManager(vCenterClient)

	view, err := manager.CreateContainerView(ctx, vCenterClient.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return
	}
	defer view.Destroy(ctx)

	var vms []mo.VirtualMachine
	fmt.Println("Retrieve list of VMs")
	err = view.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)
	if err != nil {
		return
	}
	for _, vm := range vms {
		fmt.Println(vm.Summary.Config.Name, vm.Summary.Config.GuestFullName)
	}
}

func GetVCenterClient(ctx context.Context) (*vim25.Client, error) {
	vCenterUrl := os.Getenv("VCENTER_URL")
	// TODO sanitize
	u, err := soap.ParseURL(vCenterUrl)
	if err != nil {
		return nil, err
	}

	vCenterUsername := os.Getenv("VCENTER_USERNAME")
	vCenterPassword := os.Getenv("VCENTER_PASSWORD")
	u.User = url.UserPassword(vCenterUsername, vCenterPassword)

	insecure := false
	if os.Getenv("VCENTER_INSECURE") == "1" {
		insecure = true
	}
	session := cache.Session{URL: u, Insecure: insecure}
	client := new(vim25.Client)
	fmt.Println(vCenterUrl, vCenterUsername, vCenterPassword)
	err = session.Login(ctx, client, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("Logged in")
	return client, nil
}

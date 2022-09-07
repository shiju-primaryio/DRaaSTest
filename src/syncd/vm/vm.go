package vm

import (
	// "github.com/vmware/govmomi/examples"
	// "github.com/vmware/govmomi/view"
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
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
		fmt.Println(vm.Summary.Config.Name, vm.Summary.Config.GuestFullName, vm.Summary.Config.InstanceUuid)
	}
}

func FindVmByInstanceUuid(vCenterClient *vim25.Client, instanceUuid string) (*mo.VirtualMachine, error) {
	manager := view.NewManager(vCenterClient)
	ctx := context.Background()
	view, err := manager.CreateContainerView(ctx, vCenterClient.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, err
	}
	defer view.Destroy(ctx)

	var vms []mo.VirtualMachine
	fmt.Println("Retrieve list of VMs")
	err = view.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)
	if err != nil {
		return nil, err
	}
	for _, vm := range vms {
		fmt.Println(vm.Summary.Config.Name, vm.Summary.Config.GuestFullName, vm.Summary.Config.InstanceUuid)
		if vm.Summary.Config.InstanceUuid == instanceUuid {
			fmt.Println("Found")
			return &vm, nil
		}
	}
	return nil, nil
}

// Obtain a client to call vCenter APIs
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

func CloneVm(vCenterClient *vim25.Client, vm *mo.VirtualMachine) error {
	v := object.NewVirtualMachine(vCenterClient, vm.Reference())
	ctx := context.Background()

	finder := find.NewFinder(vCenterClient)
	dataCenter, err := finder.DefaultDatacenter(ctx)
	if err != nil {
		fmt.Println("Unable to find Default Datacenter.")
		return err
	}
	fmt.Printf("dataCenter: %v\n", dataCenter)
	folders, err := dataCenter.Folders(ctx)
	if err != nil {
		fmt.Println("Unable to find folders in Datacenter.")
		return err
	}
	fmt.Printf("folders.VmFolder: %v\n", folders.VmFolder)

	dataStore, err := finder.DefaultDatastore(ctx)
	if err != nil {
		fmt.Println("Unable to find default DataStore.")
		return err
	}
	dataStoreMO := dataStore.Reference()

	resourcePool, err := finder.DefaultResourcePool(ctx)
	if err != nil {
		fmt.Println("Unable to find Default ResourcePool.")
		return err
	}

	resourcePoolMO := resourcePool.Reference()

	relocateSpec := types.VirtualMachineRelocateSpec{
		Datastore: &dataStoreMO,
		Pool:      &resourcePoolMO,
	}
	cloneSpec := &types.VirtualMachineCloneSpec{
		PowerOn:  false,
		Template: false,
	}
	cloneSpec.Location = relocateSpec
	// cloneSpec.Location.Datastore = &datastoreref
	new_vm_name := "clone-" + vm.Summary.Config.Name // TODO auto generate
	fmt.Println("Starting to clone VM")
	task, err := v.Clone(ctx, folders.VmFolder, new_vm_name, *cloneSpec)
	if err != nil {
		fmt.Printf("Unable to clone VM.")
		return err
	}
	task.Wait(ctx)
	fmt.Println("Clone task completed")
	// TODO error check
	return nil
}

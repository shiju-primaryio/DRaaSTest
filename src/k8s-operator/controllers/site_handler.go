package controllers

import (
	"context"
	"fmt"
	"net/url"
	"regexp"

	draasv1alpha1 "github.com/CacheboxInc/DRaaS/src/k8s-operator/api/v1alpha1"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func getVMList(vcenter draasv1alpha1.VCenterSpec) ([]draasv1alpha1.VMStatus, error) {
	var vmList []draasv1alpha1.VMStatus
	fmt.Println("vcenter.UserName: ", vcenter.UserName)
	fmt.Println("vcenter.Password: ", vcenter.Password)
	fmt.Println("vcenter.IP: ", vcenter.IP)
	urlString := "https://" + vcenter.UserName + ":" + vcenter.Password + "@" + vcenter.IP + "/sdk"
	fmt.Println(urlString)
	u, err := url.Parse(urlString)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect and log in to ESX or vCenter
	c, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		fmt.Println("Error connecting to ESX : ", err)
		return nil, err
	}

	f := find.NewFinder(c.Client, true)
	// Find one and only datacenter
	dc, err := f.DefaultDatacenter(ctx)
	if err != nil {
		fmt.Println("Error getting datacenter : ", err)
		return nil, err
	}

	// Make future calls local to this datacenter
	f.SetDatacenter(dc)
	vms, err := f.VirtualMachineList(ctx, "*")
	if err != nil {
		fmt.Println("Error getting VM list : ", err)
		return nil, err
	}

	pc := property.DefaultCollector(c.Client)
	// Convert datastores into list of references
	var refs []types.ManagedObjectReference
	for _, vm := range vms {
		refs = append(refs, vm.Reference())
	}

	// Retrieve all vms
	var vmt []mo.VirtualMachine
	err = pc.Retrieve(ctx, refs, nil, &vmt)
	if err != nil {
		fmt.Println("Error retrieving VM list : ", err)
		return nil, err
	}

	for _, vm := range vmt {
		vmdks := GetVmdks(vm)
		var ipAddress []string
		for _, nic := range vm.Guest.Net {
			// available in api v5
			if nic.IpConfig != nil {
				for _, addr := range nic.IpConfig.IpAddress {
					ipAddress = append(ipAddress, addr.IpAddress)
				}
			} else {
				for _, ip := range nic.IpAddress {
					ipAddress = append(ipAddress, ip)
				}
			}
		}

		vmDB := draasv1alpha1.VMStatus{
			VmId:      vm.Summary.Vm.Value,
			Name:      vm.Name,
			CPUs:      vm.Config.Hardware.NumCPU,
			MemoryMB:  vm.Config.Hardware.MemoryMB,
			GuestID:   vm.Config.GuestId,
			Uuid:      vm.Config.Uuid,
			IpAddress: ipAddress,
			NumDisks:  len(vmdks),
			Disks:     vmdks,
		}
		vmList = append(vmList, vmDB)
	}

	return vmList, nil
}

func GetVmdks(vm mo.VirtualMachine) []draasv1alpha1.Disk {
	var vmdks []draasv1alpha1.Disk
	labelPattern := regexp.MustCompile(`/[\w*-]*.vmdk$`)

	for _, device := range vm.Config.Hardware.Device {
		switch disk := device.(type) {
		case *types.VirtualDisk:
			fileName := disk.Backing.(types.BaseVirtualDeviceFileBackingInfo).
				GetVirtualDeviceFileBackingInfo().FileName
			datastore := disk.Backing.(types.BaseVirtualDeviceFileBackingInfo).
				GetVirtualDeviceFileBackingInfo().Datastore.Value
			sizeMB := disk.CapacityInKB / 1024
			thinProvisioned := *(disk.Backing.(*types.VirtualDiskFlatVer2BackingInfo).ThinProvisioned)
			unitNumber := disk.UnitNumber
			label := labelPattern.FindString(fileName)[1:]
			iofilters := disk.Iofilter

			vmdkDB := draasv1alpha1.Disk{
				Name:            fileName,
				Datastore:       datastore,
				ThinProvisioned: thinProvisioned,
				SizeMB:          sizeMB,
				UnitNumber:      *unitNumber,
				Label:           label,
				VmId:            vm.Summary.Vm.Value,
				IofilterName:    iofilters,
			}
			vmdks = append(vmdks, vmdkDB)
		}
	}

	return vmdks
}

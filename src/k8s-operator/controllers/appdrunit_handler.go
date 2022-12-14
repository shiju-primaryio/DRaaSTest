package controllers

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"

	draasv1alpha1 "github.com/CacheboxInc/DRaaS/src/k8s-operator/api/v1alpha1"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/pbm"
	pbmtypes "github.com/vmware/govmomi/pbm/types"
	"github.com/vmware/govmomi/vim25/types"
)

func ChangePolicyState(vcenter draasv1alpha1.VCenterSpec, ProtectVMUUIDList []draasv1alpha1.VmPolicyRequest) ([]draasv1alpha1.VMStatus, error) {
	var VmDetails []draasv1alpha1.VMStatus
	var vmUuid string
	var policyAttach bool

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	PolicyName := "PrimaryIO_replication"

	urlString := "https://" + vcenter.UserName + ":" + vcenter.Password + "@" + vcenter.IP + "/sdk"
	u, err := url.Parse(urlString)

	// Connect and log in to ESX or vCenter
	c, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		fmt.Println("Error connecting to ESX : ", err)
		return VmDetails, err
	}

	/*
		//Verify policy attached/detached to VM
		vmList, err := getVmList(vcenter, []string{vmUuid})
		if err != nil {
			fmt.Println("Failed to fetch VM list", err)
			return VmDetails, err
		}
	*/
	/*
		for _, vm := range vmList {
			if vm.VmUuid == vmUuid {
				VmDetails = vm
				if vm.IsProtected && policyAttach {
					fmt.Println("Policy already attached to VM.")
					err = errors.New("policy already attached to VM")
					return vm, err
				} else if !vm.IsProtected && !policyAttach {
					fmt.Println("Policy not attached to VM.")
					err = errors.New("policy not attached to VM")
					return vm, err
				}
			}
		}
	*/

	var vmUuidList []string
	for _, vmDet := range ProtectVMUUIDList {
		vmUuid = vmDet.VmUuid
		policyAttach = vmDet.IsPolicyAttach
		if policyAttach {
			vmUuidList = append(vmUuidList, vmUuid)
		}
		vmObj, err := GetVmObject(c.Client, vmUuid)
		if err != nil {
			fmt.Println("Error getting VM : ", err)
			return VmDetails, err
		}

		deviceList, err := vmObj.Device(ctx)
		if err != nil {
			fmt.Println("Error failed to fetch VM device : ", err)
			return VmDetails, err
		}

		for _, device := range deviceList {
			switch disk := device.(type) {
			case *types.VirtualDisk:
				config := &types.VirtualDeviceConfigSpec{}
				spec := types.VirtualMachineConfigSpec{}
				if policyAttach {
					policyDetails, err := GetPolicy(PolicyName, vcenter)
					if err != nil {
						fmt.Println("Unable to fetch policies from vCenter.")
						return VmDetails, err
					} else if policyDetails.PolicyId == "" {
						fmt.Println("Policy with given name not availabe at vCenter.")
						err = errors.New("policy with given name not availabe at vCenter")
						return VmDetails, err
					}

					config = &types.VirtualDeviceConfigSpec{
						Device:    disk,
						Operation: types.VirtualDeviceConfigSpecOperationEdit,
						Profile: []types.BaseVirtualMachineProfileSpec{
							&types.VirtualMachineDefinedProfileSpec{
								ProfileId: policyDetails.PolicyId,
							},
						},
					}
				} else if !policyAttach {
					config = &types.VirtualDeviceConfigSpec{
						Device:    disk,
						Operation: types.VirtualDeviceConfigSpecOperationEdit,
						Profile:   []types.BaseVirtualMachineProfileSpec{&types.VirtualMachineEmptyProfileSpec{}},
					}
				}

				spec.DeviceChange = append(spec.DeviceChange, config)
				task, err := vmObj.Reconfigure(ctx, spec)
				if err != nil {
					return VmDetails, err
				}

				err = task.Wait(ctx)
				if err != nil {
					fmt.Println("error changing disk policy: ", err)
					return VmDetails, err
				}
			}
		}
	}

	vmList, err := getVmList(vcenter, vmUuidList)
	if err != nil {
		fmt.Println("Failed to fetch VM list", err)
		return VmDetails, err
	}

	/*
		for _, vm := range vmList {
			if vm.VmUuid == vmUuid {
				fmt.Println("adding vmstatus....")
				VmDetails = vm
			}
		}
	*/
	//fmt.Println("Storage policy state changed successfully to VM : ", vmObj.Name())
	return vmList, nil
}

func GetPolicy(PolicyName string, vcenter draasv1alpha1.VCenterSpec) (draasv1alpha1.PolicyDetails, error) {
	var response_body draasv1alpha1.PolicyDetails

	policyList, err := GetPolicyList(vcenter)
	if err != nil {
		fmt.Println("Error fetching policy list from vCenter: ", err)
		return response_body, err
	}

	for _, policy := range policyList {
		if policy.PolicyName == PolicyName {
			fmt.Println("Policy available in vCenter.")
			response_body.PolicyName = policy.PolicyName
			response_body.PolicyId = policy.PolicyId
			break
		}
	}

	return response_body, err
}

func GetPolicyList(vcenter draasv1alpha1.VCenterSpec) ([]draasv1alpha1.PolicyDetails, error) {
	var policyList []draasv1alpha1.PolicyDetails
	urlString := "https://" + vcenter.UserName + ":" + vcenter.Password + "@" + vcenter.IP + "/sdk"
	u, err := url.Parse(urlString)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Connect and log in to ESX or vCenter
	c, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		fmt.Println("Error connecting to ESX : ", err)
		return nil, err
	}

	pbmSi, err := pbm.NewClient(ctx, c.Client)
	if err != nil {
		fmt.Println("Error creating pbm client : ", err)
		return nil, err
	}

	category := pbmtypes.PbmProfileCategoryEnumREQUIREMENT
	rtype := pbmtypes.PbmProfileResourceType{
		ResourceType: string(pbmtypes.PbmProfileResourceTypeEnumSTORAGE),
	}

	//Query all the profiles on the vCenter.
	ids, err := pbmSi.QueryProfile(ctx, rtype, string(category))
	if err != nil {
		fmt.Println(err)
	}

	//Retrieve the content of all profiles.
	policies, err := pbmSi.RetrieveContent(ctx, ids)
	if err != nil {
		fmt.Println(err)
	}
	var profile *pbmtypes.PbmProfile
	for i := range policies {
		profile = policies[i].GetPbmProfile()

		policy := draasv1alpha1.PolicyDetails{
			PolicyName: profile.Name,
			PolicyId:   profile.ProfileId.UniqueId,
		}

		policyList = append(policyList, policy)
	}

	return policyList, err
}

func CreateVM(vcenter draasv1alpha1.VCenterSpec, vmInfo draasv1alpha1.VMStatus) (string, error) {
	var devices object.VirtualDeviceList

	urlString := "https://" + vcenter.UserName + ":" + vcenter.Password + "@" + vcenter.IP + "/sdk"
	u, err := url.Parse(urlString)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Connect and log in to ESX or vCenter
	c, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		fmt.Println("Error connecting to ESX : ", err)
		return "", err
	}

	finder := find.NewFinder(c.Client, false)
	dc, err := finder.DefaultDatacenter(ctx)
	if err != nil {
		fmt.Println("Error finding default datacenter  : ", err)
		return "", err
	}

	finder.SetDatacenter(dc)
	ds, err := finder.DefaultDatastore(ctx)
	if err != nil {
		fmt.Println("Error finding default datacenter.  : ", err)
		return "", err
	}

	folders, err := dc.Folders(ctx)
	if err != nil {
		fmt.Println("Error finding default datacenter folders  : ", err)
		return "", err
	}

	hosts, err := finder.HostSystemList(ctx, "*/*")
	if err != nil {
		fmt.Println("Error finding host system list  : ", err)
		return "", err
	}

	nhosts := len(hosts)
	host := hosts[rand.Intn(nhosts)]
	pool, err := host.ResourcePool(ctx)
	if err != nil {
		fmt.Println("Error finding Resource Pool: ", err)
		return "", err
	}

	if nhosts == 1 {
		host = nil
	}

	vmFolder := folders.VmFolder
	var vmx string
	spec := types.VirtualMachineConfigSpec{
		// Note: real ESX allows the VM to be created without a GuestId,
		// but will power on will fail.
		Name:     vmInfo.Name,
		NumCPUs:  vmInfo.CPUs,
		MemoryMB: int64(vmInfo.MemoryMB),
		GuestId:  vmInfo.GuestID,
	}

	vmx = fmt.Sprintf("%s/%s.vmx", spec.Name, spec.Name)
	spec.Files = &types.VirtualMachineFileInfo{
		VmPathName: fmt.Sprintf("[%s] %s", ds.Name(), vmx)}

	devices, err = AddStorage(vmInfo)
	if err != nil {
		return "", err
	}

	deviceChange, err := devices.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
	if err != nil {
		return "", err
	}

	spec.DeviceChange = deviceChange
	task, cerr := vmFolder.CreateVM(ctx, spec, pool, host)
	if cerr != nil {
		fmt.Println("Error Create VM  : ", err)
		return "", err
	}

	info, err := task.WaitForResult(ctx, nil)
	if err != nil {
		fmt.Println("failed to create VM : ", cerr)
		return "", err
	}

	fmt.Println("**********Create VM succeeds************ ")
	fmt.Println("Attach Policy to disks... ")

	profileId, err := GetProfileId(c.Client)
	if err != nil {
		fmt.Println("failed to get profileId : ", cerr)
		return "", err
	}

	//Attach Storage policy to disks
	vm := object.NewVirtualMachine(c.Client, info.Result.(types.ManagedObjectReference))

	deviceList, err := vm.Device(ctx)
	for _, device := range deviceList {
		switch disk := device.(type) {
		case *types.VirtualDisk:
			spec := types.VirtualMachineConfigSpec{}
			config := &types.VirtualDeviceConfigSpec{
				Device:    disk,
				Operation: types.VirtualDeviceConfigSpecOperationEdit,
				Profile: []types.BaseVirtualMachineProfileSpec{
					&types.VirtualMachineDefinedProfileSpec{
						ProfileId: profileId,
					},
				},
			}

			spec.DeviceChange = append(spec.DeviceChange, config)

			task, err := vm.Reconfigure(ctx, spec)
			if err != nil {
				return "", err
			}

			err = task.Wait(ctx)
			if err != nil {
				fmt.Println("error changing disk policy : ", err)
				return "", err
			}
		}
	}

	return "", err
}

func AddStorage(vmInfo draasv1alpha1.VMStatus) (object.VirtualDeviceList, error) {
	var devices object.VirtualDeviceList

	for _, disk := range vmInfo.Disks {
		if vmInfo.Controller != "ide" {
			if vmInfo.Controller == "nvme" {
				nvme, err := devices.CreateNVMEController()
				if err != nil {
					return nil, err
				}

				devices = append(devices, nvme)
				vmInfo.Controller = devices.Name(nvme)
			} else {
				scsi, err := devices.CreateSCSIController("")
				if err != nil {
					return nil, err
				}

				devices = append(devices, scsi)
				vmInfo.Controller = devices.Name(scsi)
			}
		}

		//TODO
		/* // If controller is specified to be IDE or if an ISO is specified, add IDE controller.
		if vmReq.Controller == "ide" || cmd.iso != "" {
			ide, err := devices.CreateIDEController()
			if err != nil {
				return nil, err
			}

			devices = append(devices, ide)
		} */

		controller, err := devices.FindDiskController(vmInfo.Controller)
		if err != nil {
			return nil, err
		}

		disk := &types.VirtualDisk{
			VirtualDevice: types.VirtualDevice{
				Key:        devices.NewKey(),
				UnitNumber: &disk.UnitNumber,
				Backing: &types.VirtualDiskFlatVer2BackingInfo{
					DiskMode:        string(types.VirtualDiskModePersistent),
					ThinProvisioned: types.NewBool(true),
				},
			},
			CapacityInKB: disk.SizeMB * 1024,
			//TODO Iofilter: ,
		}

		devices.AssignController(disk, controller)
		devices = append(devices, disk)
	}

	return devices, nil
}

package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	draasv1alpha1 "github.com/CacheboxInc/DRaaS/src/k8s-operator/api/v1alpha1"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/vmware/govmomi/pbm"
	pbmtypes "github.com/vmware/govmomi/pbm/types"
)

/*

func getVmMap(vcenter draasv1alpha1.VCenterSpec) (map[string]draasv1alpha1.VMStatus, error) {
	vmMap := make(map[string]draasv1alpha1.VMStatus)

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
			VmId:       vm.Summary.Vm.Value,
			Name:       vm.Name,
			CPUs:       vm.Config.Hardware.NumCPU,
			MemoryMB:   vm.Config.Hardware.MemoryMB,
			GuestID:    vm.Config.GuestId,
			IpAddress:  ipAddress,
			NumDisks:   len(vmdks),
			Disks:      vmdks,
			PowerState: string(vm.Runtime.PowerState),
		}

		fmt.Println("vmDB.PowerState: ", vmDB.PowerState)

		vmMap[vm.Config.Uuid] = vmDB
	}

	return vmMap, nil
}

*/

func getVmList(vcenter draasv1alpha1.VCenterSpec, vmUuidList []string) ([]draasv1alpha1.VMStatus, error) {
	var vmList []draasv1alpha1.VMStatus
	//var vmList := make([]draasv1alpha1.VMStatus)

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
		vmdks, isProtected := GetVmdks(vm)
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
			VmId:        vm.Summary.Vm.Value,
			Name:        vm.Name,
			VmUuid:      vm.Config.Uuid,
			CPUs:        vm.Config.Hardware.NumCPU,
			MemoryMB:    vm.Config.Hardware.MemoryMB,
			GuestID:     vm.Config.GuestId,
			IpAddress:   ipAddress,
			NumDisks:    len(vmdks),
			Disks:       vmdks,
			PowerState:  string(vm.Runtime.PowerState),
			IsProtected: isProtected,
		}

		if len(vmUuidList) != 0 {
			vmList = append(vmList, vmDB)
		} else {
			for _, vmuuid := range vmUuidList {
				if vmuuid == vmDB.VmUuid {
					vmList = append(vmList, vmDB)
				}
			}
		}
	}

	vmMapNew := make([]draasv1alpha1.VMStatus, len(vmList))
	copy(vmMapNew, vmList)
	return vmMapNew, nil
}

func CreateStoragePolicyForSite(vcenter draasv1alpha1.VCenterSpec, policyDetails draasv1alpha1.StoragePolicySpec) (string, error) {

	PolicyName := "PrimaryIO_replication"
	var PolicyIdStr string
	urlString := "https://" + vcenter.UserName + ":" + vcenter.Password + "@" + vcenter.IP + "/sdk"
	u, err := url.Parse(urlString)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//TODO: Add Code to Check whether Policy exists

	// Connect and log in to ESX or vCenter
	c, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		fmt.Println("Error connecting to ESX : ", err)
		return "", err
	}

	policyinfo, err := GetPolicy(PolicyName, vcenter)
	if err != nil {
		fmt.Println("Error fetching existing policies from vCenter : ", err)
		return "", err
	}

	if policyinfo.PolicyId != "" {
		fmt.Println("Storage policy 'PrimaryIO_replication' already exists.")
		err = errors.New("storsge polivy 'PrimaryIO_replication' already exists")
		return policyinfo.PolicyId, err
	}

	pbmSi, err := pbm.NewClient(ctx, c.Client)
	if err != nil {
		fmt.Println("Error creating pbm client : ", err)
		return "", err
	}

	if policyDetails.Name == "" {
		policyDetails.Name = PolicyName
	}

	if policyDetails.Description == "" {
		policyDetails.Description = "PrimaryIO replication policy"
	}

	if policyDetails.ResourceType == "" {
		policyDetails.ResourceType = "STORAGE"
	}

	if policyDetails.Namespace == "" {
		policyDetails.Namespace = "primaryio"
	}

	if policyDetails.NamespaceId == "" {
		policyDetails.NamespaceId = "primaryio@REPLICATION"
	}

	if policyDetails.Blocksize == 0 {
		policyDetails.Blocksize = 4096
	}

	if policyDetails.Outsize == 0 {
		policyDetails.Outsize = 1024
	}

	if policyDetails.Queuesize == 0 {
		policyDetails.Queuesize = 65536
	}

	profile1 := pbmtypes.PbmCapabilityProfileCreateSpec{
		Name: policyDetails.Name,

		Description: policyDetails.Description,

		Category: string(pbmtypes.PbmProfileCategoryEnumREQUIREMENT),

		ResourceType: pbmtypes.PbmProfileResourceType{
			ResourceType: policyDetails.ResourceType,
		},

		Constraints: &pbmtypes.PbmCapabilitySubProfileConstraints{
			PbmCapabilityConstraints: pbmtypes.PbmCapabilityConstraints{},
			SubProfiles: []pbmtypes.PbmCapabilitySubProfile{
				{
					Name: "Host based services",
					Capability: []pbmtypes.PbmCapabilityInstance{
						{
							Id: pbmtypes.PbmCapabilityMetadataUniqueId{
								Namespace: policyDetails.Namespace,
								Id:        policyDetails.NamespaceId,
							},
							Constraint: []pbmtypes.PbmCapabilityConstraintInstance{
								{
									PropertyInstance: []pbmtypes.PbmCapabilityPropertyInstance{
										{
											Id:       "blocksize",
											Operator: "",
											Value:    int32(policyDetails.Blocksize),
										},
									},
								},
								{
									PropertyInstance: []pbmtypes.PbmCapabilityPropertyInstance{
										{
											Id:       "queuesize",
											Operator: "",
											Value:    int32(policyDetails.Queuesize),
										},
									},
								},
								{
									PropertyInstance: []pbmtypes.PbmCapabilityPropertyInstance{
										{
											Id:       "outsize",
											Operator: "",
											Value:    int32(policyDetails.Outsize),
										},
									},
								},
								{
									PropertyInstance: []pbmtypes.PbmCapabilityPropertyInstance{
										{
											Id:       "port",
											Operator: "",
											Value:    policyDetails.Port,
										},
									},
								},
								{
									PropertyInstance: []pbmtypes.PbmCapabilityPropertyInstance{
										{
											Id:       "user",
											Operator: "",
											Value:    policyDetails.User,
										},
									},
								},
								{
									PropertyInstance: []pbmtypes.PbmCapabilityPropertyInstance{
										{
											Id:       "secret",
											Operator: "",
											Value:    policyDetails.Secret,
										},
									},
								},
								{
									PropertyInstance: []pbmtypes.PbmCapabilityPropertyInstance{
										{
											Id:       "host",
											Operator: "",
											Value:    policyDetails.Host,
										},
									},
								},
								{
									PropertyInstance: []pbmtypes.PbmCapabilityPropertyInstance{
										{
											Id:       "enckey",
											Operator: "",
											Value:    policyDetails.Enckey,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	id, err := pbmSi.CreateProfile(ctx, profile1)
	if err != nil {
		fmt.Println("Error creating profile : ", err)
		return "", err
	}
	PolicyIdStr = id.UniqueId
	fmt.Println("*******Policy creation succeds*****")

	fmt.Println("Policy ID: ", id.UniqueId)
	return PolicyIdStr, err
}

func DeleteStoragePolicy(vcenter draasv1alpha1.VCenterSpec, policyDetails draasv1alpha1.StoragePolicySpec) error {
	PolicyName := "PrimaryIO_replication"
	urlString := "https://" + vcenter.UserName + ":" + vcenter.Password + "@" + vcenter.IP + "/sdk"
	u, err := url.Parse(urlString)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect and log in to ESX or vCenter
	c, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		fmt.Println("Error connecting to ESX : ", err)
		return err
	}

	policyinfo, err := GetPolicy(PolicyName, vcenter)
	if err != nil {
		fmt.Println("Error fetching existing policies from vCenter : ", err)
		return err
	}

	if policyinfo.PolicyId == "" {
		fmt.Println("Storage policy 'PrimaryIO_replication' not available at vCenter.")
		err = errors.New("storsge polivy 'PrimaryIO_replication' not available at vCenter")
		return err
	}

	pbmSi, err := pbm.NewClient(ctx, c.Client)
	if err != nil {
		fmt.Println("Error creating pbm client : ", err)
		return err
	}

	var profileIds []pbmtypes.PbmProfileId
	profileIds = append(profileIds, pbmtypes.PbmProfileId{UniqueId: policyinfo.PolicyId})

	//delete profile
	outcome, err := pbmSi.DeleteProfile(ctx, profileIds)
	if err != nil {
		fmt.Println("Error delete policy : ", err)
		return err
	}

	fmt.Println(outcome)
	return nil

}

func GetVmdks(vm mo.VirtualMachine) ([]draasv1alpha1.Disk, bool) {
	var vmdks []draasv1alpha1.Disk
	var isProtected bool
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
			var filterName string
			//iofilters := disk.Iofilter
			for _, iof := range disk.Iofilter {
				if strings.Contains(iof, "primaryio") {
					filterName = iof
					isProtected = true
					break
				}
			}

			//fmt.Println("--- filterName: ", filterName)

			vmdkDB := draasv1alpha1.Disk{
				Name:            fileName,
				Datastore:       datastore,
				IofilterName:    filterName,
				ThinProvisioned: thinProvisioned,
				SizeMB:          sizeMB,
				UnitNumber:      *unitNumber,
				Label:           label,
				VmId:            vm.Summary.Vm.Value,
			}
			vmdks = append(vmdks, vmdkDB)
		}
	}

	return vmdks, isProtected
}

func VmPowerChange(vcenter draasv1alpha1.VCenterSpec, vMuuid string, powerState bool) (string, error) {
	urlString := "https://" + vcenter.UserName + ":" + vcenter.Password + "@" + vcenter.IP + "/sdk"
	u, err := url.Parse(urlString)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect and log in to ESX or vCenter
	c, err1 := govmomi.NewClient(ctx, u, true)
	if err1 != nil {
		fmt.Println("Error connecting to ESX : ", err1)
		return "", err1
	}

	f := find.NewFinder(c.Client, true)
	// Find one and only datacenter
	dc, err := f.DefaultDatacenter(ctx)
	if err != nil {
		fmt.Println("Error getting datacenter : ", err)
		return "", err
	}

	pc := property.DefaultCollector(c.Client)
	s := object.NewSearchIndex(c.Client)
	svm, err := s.FindByUuid(context.Background(), dc, vMuuid, true, nil)
	if err != nil {
		fmt.Println("Error retrieving VM Object : ", err)
		return "", err
	}

	fmt.Println("UUID after get :", svm.Reference())
	var vm mo.VirtualMachine
	err = pc.RetrieveOne(ctx, svm.Reference(), nil, &vm)
	if err != nil {
		fmt.Println("Error retrieving VM after VM: ", err)
		return "", err
	}
	var task *object.Task

	vmObj := object.NewVirtualMachine(c.Client, vm.Reference())

	CurrentPowerState, _ := vmObj.PowerState(ctx)

	if (powerState) && (CurrentPowerState == types.VirtualMachinePowerStatePoweredOff) {
		fmt.Println("Powering on VM...")
		task, err = vmObj.PowerOn(ctx)
	} else if (!powerState) && (CurrentPowerState == types.VirtualMachinePowerStatePoweredOn) {
		fmt.Println("Powering off VM...")
		task, err = vmObj.PowerOff(ctx)
	} else {
		fmt.Println("Already in the desired power state. Not doing anything.")
		return "", nil
	}

	if err != nil {
		fmt.Printf("Failed to change power state of VM.")
		//return "", err
	}

	info, err := task.WaitForResult(ctx, nil)
	if err != nil {
		fmt.Printf("VM change power state process failed: %v", err)
		return "", err
	}

	//fmt.Println("Power change task: ", info.State)
	return info.Name, err
}

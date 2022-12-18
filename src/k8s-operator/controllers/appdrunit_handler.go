package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

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
			fmt.Println("Trying to attach policy for vmuuid:  ", vmUuid)
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

		bPolicyIsAlreadyAttached := false
		for _, device := range deviceList {
			switch disk := device.(type) {
			case *types.VirtualDisk:
				var vmProfilespec []types.BaseVirtualMachineProfileSpec
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
					for _, iof := range disk.Iofilter {
						if strings.Contains(iof, "primaryio") {
							fmt.Println("Policy is already attached to VMDK of vm:", vmObj.Name())
							bPolicyIsAlreadyAttached = true
							break
						}
					}
					if bPolicyIsAlreadyAttached {
						break
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
					vmProfilespec = []types.BaseVirtualMachineProfileSpec{
						&types.VirtualMachineDefinedProfileSpec{
							ProfileId: policyDetails.PolicyId,
						},
					}
				} else if !policyAttach {
					config = &types.VirtualDeviceConfigSpec{
						Device:    disk,
						Operation: types.VirtualDeviceConfigSpecOperationEdit,
						Profile:   []types.BaseVirtualMachineProfileSpec{&types.VirtualMachineEmptyProfileSpec{}},
					}
					vmProfilespec = []types.BaseVirtualMachineProfileSpec{&types.VirtualMachineEmptyProfileSpec{}}
				}

				spec.DeviceChange = append(spec.DeviceChange, config)
				spec.VmProfile = vmProfilespec
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

	if len(vmUuidList) != 0 {
		vmList, err := getVmList(vcenter, vmUuidList)
		if err != nil {
			fmt.Println("Failed to fetch VM list", err)
			return VmDetails, err
		}
		return vmList, nil
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
	return VmDetails, nil
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

func GetVMDKsFromPostGresDB(VesAuthToken string) (draasv1alpha1.VMDKListFromPostGresDResponse, error) {
	//vesauth, _ := ctx.Request.Cookie("VESauth")
	var result draasv1alpha1.VMDKListFromPostGresDResponse

	url2 := "https://r81d6d155168c.snif-d060ea6e909e-9c3701e2.snif.xyz/api/vmdks"
	req2, _ := http.NewRequest("GET", url2, nil)
	req2.Header.Add("content-type", "application/json")
	req2.Header.Add("cache-control", "no-cache")
	req2.Header.Add("X-VES-Authorization", VesAuthToken)

	//fmt.Println("\nRequest PHP API URL", url2)

	//fmt.Println("\nRequest PHP API", req2)

	res2, err2 := http.DefaultClient.Do(req2)
	if err2 != nil {
		fmt.Println(err2)
	} else {
		defer res2.Body.Close()
		body2, _ := ioutil.ReadAll(res2.Body)
		if err := json.Unmarshal(body2, &result); err != nil { // Parse []byte to the go struct pointer
			fmt.Println(err)
			fmt.Println("Can not unmarshal JSON")
			return result, err
		}
		//fmt.Println(result.Data)
	}
	return result, nil

}

type FailoverResponse struct {
	Data struct {
		Id           string `json:"id"`
		SourceVMDKId string `json:"vmdk_id"`
		TargetVMDKId string `json:"new_vmdk_id"`
		Ack          string `json:"ack"`
		Active       bool   `json:"active"`
		Sent_ct      string `json:"sent_ct"`
		Sentblocks   string `json:"sentblocks"`
		TotalBlocks  string `json:"totalBlocks"`
	} `json:"data"`
}

func InitiateFailover(VesAuthToken string, vmdkmapList []draasv1alpha1.TriggerFailoverVmdkMapping) error {

	for i, vmdkmap := range vmdkmapList {

		fmt.Println("\tInitiateFailover: vmdkmap.SourceVmdkID :", vmdkmap.SourceVmdkID)
		fmt.Println("\tInitiateFailover: vmdkmap.TargetVmdkID :", vmdkmap.TargetVmdkID)

		if (vmdkmap.SourceVmdkID == "") || (vmdkmap.TargetVmdkID == "") {
			fmt.Println("Continuing")
			continue
		}
		//vesauth, _ := ctx.Request.Cookie("VESauth")
		url2 := "https://r81d6d155168c.snif-d060ea6e909e-9c3701e2.snif.xyz/api/failovers"

		//var jsonStr = []byte(`{"vmdk_id":"56", "new_vmdk_id":"77"}`)
		jsonData := map[string]string{"vmdk_id": vmdkmap.SourceVmdkID, "new_vmdk_id": vmdkmap.TargetVmdkID}
		jsonStr, _ := json.Marshal(jsonData)

		req2, _ := http.NewRequest("POST", url2, bytes.NewBuffer(jsonStr))
		req2.Header.Add("content-type", "application/json")
		req2.Header.Add("cache-control", "no-cache")
		req2.Header.Add("X-VES-Authorization", VesAuthToken)

		fmt.Println("Request PHP API", req2)

		res2, err2 := http.DefaultClient.Do(req2)
		if err2 != nil {
			fmt.Println(err2)
		} else {
			defer res2.Body.Close()
			body2, _ := ioutil.ReadAll(res2.Body)
			var result FailoverResponse
			if err := json.Unmarshal(body2, &result); err != nil { // Parse []byte to the go struct pointer
				fmt.Println(err)
				fmt.Println("Can not unmarshal JSON")
			}
			fmt.Println("Failover Id (vmdkmap.FailoverTriggerID) created by Failover API", result.Data.Id)
			vmdkmap.FailoverTriggerID = result.Data.Id
			vmdkmap.Ack = result.Data.Ack
			vmdkmap.ActiveFailover = result.Data.Active
			vmdkmap.SentCT = result.Data.Sent_ct
			vmdkmap.SentBlocks = result.Data.Sentblocks
			vmdkmap.TotalBlocks = result.Data.TotalBlocks

			vmdkmapList[i] = vmdkmap
		}
	}
	return nil

}

func WaitForActiveBitTobeSet(VesAuthToken string, vmdkmapList []draasv1alpha1.TriggerFailoverVmdkMapping) error {
	var bRetryActiveBit bool

	RetryCount := 0
	bRetryActiveBit = true
	for bRetryActiveBit {

		fmt.Println("Sleep Over for 5 seconds in WaitForActiveBitTobeSet.....")
		// Calling Sleep method
		time.Sleep(5 * time.Second)

		for i, vmdkmap := range vmdkmapList {

			fmt.Println("\t WaitForActiveBitTobeSet: vmdkmap.SourceVmdkID :", vmdkmap.SourceVmdkID)
			fmt.Println("\t WaitForActiveBitTobeSet: vmdkmap.TargetVmdkID :", vmdkmap.TargetVmdkID)
			fmt.Println("\t WaitForActiveBitTobeSet: vmdkmap.TargetVmdkID :", vmdkmap.ActiveFailover)

			if (vmdkmap.SourceVmdkID == "") || (vmdkmap.TargetVmdkID == "") || (vmdkmap.ActiveFailover == true) {
				RetryCount = RetryCount + 1
				if RetryCount > 5 {
					//TODO: Revisit this code
					bRetryActiveBit = false
					fmt.Println("Failover API:exceeded retry count of 5 exiting : ")
					return nil
					//continue
				}
				fmt.Println("Continuing")
				continue
			}
			//FailoverId string
			FailoverId := vmdkmap.FailoverTriggerID
			//vesauth, _ := ctx.Request.Cookie("VESauth")
			url2 := "https://r81d6d155168c.snif-d060ea6e909e-9c3701e2.snif.xyz/api/failovers/"

			url2 += FailoverId
			//var jsonStr = []byte(`{"vmdk_id":"56", "new_vmdk_id":"77"}`)
			//jsonData := map[string]string{"vmdk_id": vmdkmap.SourceVmdkID, "new_vmdk_id": vmdkmap.TargetVmdkID}
			//jsonStr, _ := json.Marshal(jsonData)

			req2, _ := http.NewRequest("GET", url2, nil)
			req2.Header.Add("content-type", "application/json")
			req2.Header.Add("cache-control", "no-cache")
			req2.Header.Add("X-VES-Authorization", VesAuthToken)

			fmt.Println("Request PHP API", req2)

			res2, err2 := http.DefaultClient.Do(req2)
			if err2 != nil {
				fmt.Println(err2)
			} else {
				defer res2.Body.Close()
				body2, _ := ioutil.ReadAll(res2.Body)
				var result FailoverResponse
				if err := json.Unmarshal(body2, &result); err != nil { // Parse []byte to the go struct pointer
					fmt.Println(err)
					fmt.Println("Can not unmarshal JSON")
				}
				fmt.Println("WaitForActiveBitTobeSet: Failover API: Failover Id : ", result.Data.Id)
				//vmdkmap.FailoverTriggerID = result.Data.Id
				vmdkmap.Ack = result.Data.Ack
				vmdkmap.ActiveFailover = result.Data.Active
				if !vmdkmap.ActiveFailover {
					bRetryActiveBit = false
					fmt.Println("Failover API: Active bit is false for failover ID : ", result.Data.Id)
				}
				fmt.Println("WaitForActiveBitTobeSet: Failover API: Failover Sent Blocks : ", result.Data.Sentblocks)
				vmdkmap.SentCT = result.Data.Sent_ct
				vmdkmap.SentBlocks = result.Data.Sentblocks
				vmdkmap.TotalBlocks = result.Data.TotalBlocks

				vmdkmapList[i] = vmdkmap
			}
		}
	}
	return nil
}

func GetFailoverStatus(VesAuthToken string, vmdkmapList []draasv1alpha1.TriggerFailoverVmdkMapping) error {

	for i, vmdkmap := range vmdkmapList {

		fmt.Println("\t GetFailoverStatus: vmdkmap.SourceVmdkID :", vmdkmap.SourceVmdkID)
		fmt.Println("\t GetFailoverStatus: vmdkmap.TargetVmdkID :", vmdkmap.TargetVmdkID)

		if (vmdkmap.SourceVmdkID == "") || (vmdkmap.TargetVmdkID == "") {
			fmt.Println("Continuing")
			continue
		}
		//FailoverId string
		FailoverId := vmdkmap.FailoverTriggerID
		//vesauth, _ := ctx.Request.Cookie("VESauth")
		url2 := "https://r81d6d155168c.snif-d060ea6e909e-9c3701e2.snif.xyz/api/failovers/"

		url2 += FailoverId
		//var jsonStr = []byte(`{"vmdk_id":"56", "new_vmdk_id":"77"}`)
		//jsonData := map[string]string{"vmdk_id": vmdkmap.SourceVmdkID, "new_vmdk_id": vmdkmap.TargetVmdkID}
		//jsonStr, _ := json.Marshal(jsonData)

		req2, _ := http.NewRequest("GET", url2, nil)
		req2.Header.Add("content-type", "application/json")
		req2.Header.Add("cache-control", "no-cache")
		req2.Header.Add("X-VES-Authorization", VesAuthToken)

		fmt.Println("Failover status url: ", url2)
		fmt.Println("Request PHP API", req2)

		res2, err2 := http.DefaultClient.Do(req2)
		if err2 != nil {
			fmt.Println(err2)
		} else {
			defer res2.Body.Close()
			body2, _ := ioutil.ReadAll(res2.Body)
			var result FailoverResponse
			if err := json.Unmarshal(body2, &result); err != nil { // Parse []byte to the go struct pointer
				fmt.Println(err)
				fmt.Println("Can not unmarshal JSON")
			}
			fmt.Println("GET Failover API: Failover Id : ", result.Data.Id)
			//vmdkmap.FailoverTriggerID = result.Data.Id
			fmt.Println("Failover API: Failover ACk : ", result.Data.Ack)
			vmdkmap.Ack = result.Data.Ack
			fmt.Println("GET Failover API: Failover Active flag : ", result.Data.Active)
			vmdkmap.ActiveFailover = result.Data.Active
			fmt.Println("GET Failover API: Failover Sent Blocks : ", result.Data.Sentblocks)
			vmdkmap.SentCT = result.Data.Sent_ct
			vmdkmap.SentBlocks = result.Data.Sentblocks
			vmdkmap.TotalBlocks = result.Data.TotalBlocks
			fmt.Println("GET Failover API: Failover Total Blocks : ", result.Data.TotalBlocks)

			vmdkmapList[i] = vmdkmap
		}
	}

	return nil

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
			spec.VmProfile = []types.BaseVirtualMachineProfileSpec{
				&types.VirtualMachineDefinedProfileSpec{
					ProfileId: profileId,
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

	/*
		fmt.Println("Powering ON VM... ")

		// PowerON VM
		task, err = vm.PowerOn(ctx)
		if err != nil {
			fmt.Printf("Failed to change power state of VM.")
			//return "", err
		}

		fmt.Printf("Sleeping for 2 Seconds...")
		time.Sleep(2 * time.Second)

		_, err1 := task.WaitForResult(ctx, nil)
		if err1 != nil {
			fmt.Printf("VM change power state process failed: %v", err)
			return "", err
		}

		fmt.Println("Powering OFF VM... ")

		// PowerOFF VM
		task, err = vm.PowerOff(ctx)
		if err != nil {
			fmt.Printf("Failed to change power state of VM.")
			//return "", err
		}

		_, err1 = task.WaitForResult(ctx, nil)
		if err1 != nil {
			fmt.Printf("VM change power state process failed: %v", err)
			return "", err
		}

	*/
	return "", err
}

func AddStorage(vmInfo draasv1alpha1.VMStatus) (object.VirtualDeviceList, error) {
	var devices object.VirtualDeviceList

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

	for _, disk := range vmInfo.Disks {
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

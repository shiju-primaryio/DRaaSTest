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

// Alpha Environment  PHP Server
const SnifPhpUrl string = "https://rea93e992fa2f.snif-0e92f727f614-81cbba95.snif.xyz"

//Dev Environment PHP Server
//const SnifPhpUrl string = "https://rcf76eb14e093.snif-3ba203b1de68-4801cd35.snif.xyz"

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

	url2 := SnifPhpUrl + "/api/vmdks"
	req2, _ := http.NewRequest("GET", url2, nil)
	req2.Header.Add("content-type", "application/json")
	req2.Header.Add("cache-control", "no-cache")
	req2.Header.Add("X-VES-Authorization", VesAuthToken)

	//fmt.Println("\nRequest PHP API URL", url2)

	//fmt.Println("\nRequest PHP API", req2)
	//skip ssl tls verify
	//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

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

func CreateVMDKsAtPostGresDB(VesAuthToken string, vmInfo draasv1alpha1.VMStatus) error {
	fmt.Println("In CreateVMDKsAtPostGresDB.....")
	fmt.Println("SnifPhpUrl : ", SnifPhpUrl)

	url2 := SnifPhpUrl + "/api/vmdks"
	fmt.Println("url2 : ", url2)

	for _, disk := range vmInfo.Disks {
		fmt.Println("disk.AbsPath: ", disk.AbsPath)
		var result draasv1alpha1.VMDKFromPostGresDResponse

		jsonData := map[string]string{"scope": disk.AbsPath}
		jsonStr, _ := json.Marshal(jsonData)

		req2, _ := http.NewRequest("POST", url2, bytes.NewBuffer(jsonStr))
		req2.Header.Add("content-type", "application/json")
		req2.Header.Add("cache-control", "no-cache")
		req2.Header.Add("X-VES-Authorization", VesAuthToken)

		fmt.Println("CreateVMDK Request PHP API", req2)
		res2, err2 := http.DefaultClient.Do(req2)
		if err2 != nil {
			fmt.Println(err2)
		} else {
			defer res2.Body.Close()
			body2, _ := ioutil.ReadAll(res2.Body)
			if err := json.Unmarshal(body2, &result); err != nil { // Parse []byte to the go struct pointer
				fmt.Println(err)
				fmt.Println("Can not unmarshal JSON")
				return err
			}

			fmt.Println("result.VmdkId: ", result.VmdkId)
			//disk.VmdkPostgerId = result.VmdkId
		}
	}

	return nil
}

func DeleteFailoverEntry(VesAuthToken string, FailOverOrFailBackId string) (bool, error) {

	url := SnifPhpUrl + "/api/failovers/" + FailOverOrFailBackId

	req2, _ := http.NewRequest("DELETE", url, nil)
	req2.Header.Add("content-type", "application/json")
	req2.Header.Add("cache-control", "no-cache")
	req2.Header.Add("X-VES-Authorization", VesAuthToken)

	//fmt.Println("CancelFailover Request PHP API", req2)

	res2, err2 := http.DefaultClient.Do(req2)
	if err2 != nil {
		fmt.Println("Cancel failover failed for id: ", FailOverOrFailBackId)
		fmt.Println(err2)
		return false, err2
	}

	if res2.StatusCode != 204 {
		fmt.Println("Cancel failover failed for id: ", FailOverOrFailBackId)
	}

	fmt.Println("Cancel failover succeds for id: ", FailOverOrFailBackId)
	return true, nil
}

func CancelFailover(VesAuthToken string, vmmapList []draasv1alpha1.TriggerFailoverVmMapping) error {
	//fmt.Println("In CancelFailover.....")
	//url := SnifPhpUrl + "/api/failovers"

	for _, vmmap := range vmmapList {
		for _, vmdkmap := range vmmap.VmdkStatusList {
			fmt.Println("\tCancelFailover: vmmap.FailoverTriggerID", vmdkmap.FailoverTriggerID)
			//for _,vmdk : =
			DeleteFailoverEntry(VesAuthToken, vmdkmap.FailoverTriggerID)
		}
	}
	return nil
}

func InitiateFailover(VesAuthToken string, vmmapList []draasv1alpha1.TriggerFailoverVmMapping) error {

	for i, vmmap := range vmmapList {

		for j, vmdkmap := range vmmap.VmdkStatusList {

			fmt.Println("\tInitiateFailover: vmdkmap.SourceVmdkID :", vmdkmap.SourceVmdkID)
			fmt.Println("\tInitiateFailover: vmdkmap.TargetVmdkID :", vmdkmap.TargetVmdkID)

			if (vmdkmap.SourceVmdkID == "") || (vmdkmap.TargetVmdkID == "") {
				fmt.Println("Continuing")
				continue
			}

			if vmdkmap.FailoverTriggerID != "" {
				fmt.Println("Failover is already initiated for Vm : ", vmdkmap.Label)
				continue
			}

			//vesauth, _ := ctx.Request.Cookie("VESauth")
			url2 := SnifPhpUrl + "/api/failovers"

			//var jsonStr = []byte(`{"vmdk_id":"56", "new_vmdk_id":"77"}`)
			jsonData := map[string]string{"vmdk_id": vmdkmap.SourceVmdkID, "new_vmdk_id": vmdkmap.TargetVmdkID}
			jsonStr, _ := json.Marshal(jsonData)

			//skip ssl tls verify
			//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

			req2, _ := http.NewRequest("POST", url2, bytes.NewBuffer(jsonStr))
			req2.Header.Add("content-type", "application/json")
			req2.Header.Add("cache-control", "no-cache")
			req2.Header.Add("X-VES-Authorization", VesAuthToken)

			//fmt.Println("Request PHP API", req2)

			res2, err2 := http.DefaultClient.Do(req2)
			if err2 != nil {
				fmt.Println(err2)
			} else {
				defer res2.Body.Close()
				body2, _ := ioutil.ReadAll(res2.Body)
				var result draasv1alpha1.FailoverResponse
				if err := json.Unmarshal(body2, &result); err != nil { // Parse []byte to the go struct pointer
					fmt.Println(err)
					fmt.Println("Can not unmarshal JSON")
				}
				fmt.Println("Failover Id (vmdkmap.FailoverTriggerID) created by Failover API", result.Data.Id)
				vmmapList[i].VmdkStatusList[j].FailoverTriggerID = result.Data.Id
				vmmapList[i].VmdkStatusList[j].Ack = result.Data.Ack
				vmmapList[i].VmdkStatusList[j].ActiveFailover = result.Data.Active
				vmmapList[i].VmdkStatusList[j].SentCT = result.Data.Sent_ct
				vmmapList[i].VmdkStatusList[j].SentBlocks = result.Data.Sentblocks
				vmmapList[i].VmdkStatusList[j].TotalBlocks = result.Data.TotalBlocks
			}
		}
	}
	return nil
}

func InitiateFailback(VesAuthToken string, bFailbackWithLiveReplication bool, vmmapList []draasv1alpha1.TriggerFailbackVmMapping) error {

	for i, vmmap := range vmmapList {
		fmt.Println("\tInitiateFailback: vmdkmap.vmname :", vmmap.VmName)

		for j, vmdkmap := range vmmap.VmdkStatusList {

			fmt.Println("\tInitiateFailback: vmdkmap.SourceVmdkID :", vmdkmap.SourceVmdkID)
			fmt.Println("\tInitiateFailback: vmdkmap.TargetVmdkID :", vmdkmap.TargetVmdkID)
			if (vmdkmap.SourceVmdkID == "") || (vmdkmap.TargetVmdkID == "") {
				fmt.Println("Continuing")
				continue
			}

			if vmdkmap.FailbackTriggerID != "" {
				fmt.Println("Failback is already initiated for Vm : ", vmdkmap.Label)
				continue
			}

			//vesauth, _ := ctx.Request.Cookie("VESauth")
			url2 := SnifPhpUrl + "/api/failovers"

			//var jsonStr = []byte(`{"vmdk_id":"56", "new_vmdk_id":"77"}`)
			jsonData := draasv1alpha1.InitiateFailbackRequest{
				SourceVMDKId: vmdkmap.SourceVmdkID,
				TargetVMDKId: vmdkmap.TargetVmdkID,
				Follow:       bFailbackWithLiveReplication,
			}

			//map[string]string{"vmdk_id": vmdkmap.SourceVmdkID, "new_vmdk_id": vmdkmap.TargetVmdkID, "follow": bFailbackWithLiveReplication}
			jsonStr, _ := json.Marshal(jsonData)

			//skip ssl tls verify
			//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

			req2, _ := http.NewRequest("POST", url2, bytes.NewBuffer(jsonStr))
			//fmt.Println("Request PHP API URL", req2)
			req2.Header.Add("content-type", "application/json")
			req2.Header.Add("cache-control", "no-cache")
			req2.Header.Add("X-VES-Authorization", VesAuthToken)

			//fmt.Println("Request PHP API", req2)

			res2, err2 := http.DefaultClient.Do(req2)
			if err2 != nil {
				fmt.Println(err2)
			} else {
				defer res2.Body.Close()
				body2, _ := ioutil.ReadAll(res2.Body)
				var result draasv1alpha1.FailoverResponse
				if err := json.Unmarshal(body2, &result); err != nil { // Parse []byte to the go struct pointer
					fmt.Println(err)
					fmt.Println("Can not unmarshal JSON")
				}
				fmt.Println("Failover Id (vmdkmap.FailoverTriggerID) created by Failover API", result.Data.Id)
				vmmapList[i].VmdkStatusList[j].FailbackTriggerID = result.Data.Id
				vmmapList[i].VmdkStatusList[j].Ack = result.Data.Ack
				vmmapList[i].VmdkStatusList[j].ActiveFailback = result.Data.Active
				vmmapList[i].VmdkStatusList[j].SentCT = result.Data.Sent_ct
				vmmapList[i].VmdkStatusList[j].SentBlocks = result.Data.Sentblocks
				vmmapList[i].VmdkStatusList[j].TotalBlocks = result.Data.TotalBlocks
				vmmapList[i].VmdkStatusList[j].Follow_Seq = result.Data.Follow_Seq

				//vmdkmapList[i] = vmdkmap
			}
		}
	}
	return nil

}

func WaitForActiveBitTobeSetFailBack(VesAuthToken string, vmmapList []draasv1alpha1.TriggerFailbackVmMapping) error {

	MaxRetryChecks := 50

	for i, vmmap := range vmmapList {

		//Setting Active Bit for VM is True. If One of the vmdk's active bit is false, then it resets it.
		vmmapList[i].IsActiveBitTrue = true

		for j, vmdkmap := range vmmap.VmdkStatusList {

			fmt.Println("\t WaitForActiveBitTobeSetFailBack: vmdkmap.SourceVmdkID :", vmdkmap.SourceVmdkID)
			fmt.Println("\t WaitForActiveBitTobeSetFailBack: vmdkmap.TargetVmdkID :", vmdkmap.TargetVmdkID)
			fmt.Println("\t WaitForActiveBitTobeSetFailBack: vmdkmap.ActiveFailover :", vmdkmap.ActiveFailback)

			if (vmdkmap.SourceVmdkID == "") || (vmdkmap.TargetVmdkID == "") || (vmdkmap.ActiveFailback == true) {
				continue
			}

			// This is just start of failover. Wait is needed for active bit to be set
			for k := 0; k < MaxRetryChecks; k++ {

				//FailbackId string
				FailbackId := vmdkmap.FailbackTriggerID
				//vesauth, _ := ctx.Request.Cookie("VESauth")
				url2 := SnifPhpUrl + "/api/failovers/"

				url2 += FailbackId

				req2, _ := http.NewRequest("GET", url2, nil)
				req2.Header.Add("content-type", "application/json")
				req2.Header.Add("cache-control", "no-cache")
				req2.Header.Add("X-VES-Authorization", VesAuthToken)

				//fmt.Println("Request PHP API", req2)

				//skip ssl tls verify
				//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

				res2, err2 := http.DefaultClient.Do(req2)
				if err2 != nil {
					fmt.Println(err2)
				} else {
					defer res2.Body.Close()
					body2, _ := ioutil.ReadAll(res2.Body)
					var result draasv1alpha1.FailoverResponse
					if err := json.Unmarshal(body2, &result); err != nil { // Parse []byte to the go struct pointer
						fmt.Println(err)
						fmt.Println("Can not unmarshal JSON")
					}
					fmt.Println("WaitForActiveBitTobeSetFailBack: Failback API: Failback Id : ", result.Data.Id)
					//vmdkmap.FailoverTriggerID = result.Data.Id
					vmmapList[i].VmdkStatusList[j].Ack = result.Data.Ack
					vmmapList[i].VmdkStatusList[j].ActiveFailback = result.Data.Active

					fmt.Println("WaitForActiveBitTobeSetFailBack: Failover API: Failback Sent Blocks : ", result.Data.Sentblocks)
					vmmapList[i].VmdkStatusList[j].SentCT = result.Data.Sent_ct
					vmmapList[i].VmdkStatusList[j].SentBlocks = result.Data.Sentblocks
					vmmapList[i].VmdkStatusList[j].TotalBlocks = result.Data.TotalBlocks

					if vmdkmap.ActiveFailback == false {
						vmmapList[i].IsActiveBitTrue = vmdkmap.ActiveFailback
						fmt.Println("Failback API:Sleeping for 5 seconds for active bit to be true")
						time.Sleep(5 * time.Second)
					} else {
						fmt.Println("Failback API: Active bit is true for failback ID : ", result.Data.Id)
						break
					}
				}
			}
		}
		vmmapList[i].IsActiveBitTrue = true
		for _, vmdkmap := range vmmap.VmdkStatusList {
			if vmdkmap.ActiveFailback == false {
				vmmapList[i].IsActiveBitTrue = false
			}
		}
	}
	return nil
}

func WaitForActiveBitTobeSet(VesAuthToken string, vmmapList []draasv1alpha1.TriggerFailoverVmMapping) error {

	MaxRetryChecks := 50

	for i, vmmap := range vmmapList {

		//Setting Active Bit for VM is True. If One of the vmdk's active bit is false, then it resets it.
		vmmapList[i].IsActiveBitTrue = true

		for j, vmdkmap := range vmmap.VmdkStatusList {

			fmt.Println("\t WaitForActiveBitTobeSet: vmdkmap.SourceVmdkID :", vmdkmap.SourceVmdkID)
			fmt.Println("\t WaitForActiveBitTobeSet: vmdkmap.TargetVmdkID :", vmdkmap.TargetVmdkID)
			fmt.Println("\t WaitForActiveBitTobeSet: vmdkmap.ActiveFailover :", vmdkmap.ActiveFailover)

			if (vmdkmap.SourceVmdkID == "") || (vmdkmap.TargetVmdkID == "") || (vmdkmap.ActiveFailover == true) {
				continue
			}

			// This is just start of failover. Wait is needed for active bit to be set
			for k := 0; k < MaxRetryChecks; k++ {

				//FailbackId string
				FailOverId := vmdkmap.FailoverTriggerID
				//vesauth, _ := ctx.Request.Cookie("VESauth")
				url2 := SnifPhpUrl + "/api/failovers/"

				url2 += FailOverId

				req2, _ := http.NewRequest("GET", url2, nil)
				req2.Header.Add("content-type", "application/json")
				req2.Header.Add("cache-control", "no-cache")
				req2.Header.Add("X-VES-Authorization", VesAuthToken)

				//fmt.Println("Request PHP API", req2)

				//skip ssl tls verify
				//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

				res2, err2 := http.DefaultClient.Do(req2)
				if err2 != nil {
					fmt.Println(err2)
				} else {
					defer res2.Body.Close()
					body2, _ := ioutil.ReadAll(res2.Body)
					var result draasv1alpha1.FailoverResponse
					if err := json.Unmarshal(body2, &result); err != nil { // Parse []byte to the go struct pointer
						fmt.Println(err)
						fmt.Println("Can not unmarshal JSON")
					}
					fmt.Println("WaitForActiveBitTobeSet: FailOver API: FailOver Id : ", result.Data.Id)
					//vmdkmap.FailoverTriggerID = result.Data.Id
					vmmapList[i].VmdkStatusList[j].Ack = result.Data.Ack
					vmmapList[i].VmdkStatusList[j].ActiveFailover = result.Data.Active
					if !vmdkmap.ActiveFailover {
						fmt.Println("Failover API: Active bit is false for failover ID : ", result.Data.Id)
						//Setting to false
						vmmapList[i].IsActiveBitTrue = vmdkmap.ActiveFailover
					}
					fmt.Println("WaitForActiveBitTobeSet: Failover API: Failover Sent Blocks : ", result.Data.Sentblocks)
					vmmapList[i].VmdkStatusList[j].SentCT = result.Data.Sent_ct
					vmmapList[i].VmdkStatusList[j].SentBlocks = result.Data.Sentblocks
					vmmapList[i].VmdkStatusList[j].TotalBlocks = result.Data.TotalBlocks
					//vmdkmapList[i] = vmdkmap
					if vmdkmap.ActiveFailover == false {
						vmmapList[i].IsActiveBitTrue = vmdkmap.ActiveFailover
						fmt.Println("Failover API:Sleeping for 5 seconds for active bit to be true")
						time.Sleep(5 * time.Second)
					} else {
						fmt.Println("Failover API: Active bit is true for failover ID : ", result.Data.Id)
						break
					}
				}
			}
		}
		vmmapList[i].IsActiveBitTrue = true
		for _, vmdkmap := range vmmap.VmdkStatusList {
			if vmdkmap.ActiveFailover == false {
				vmmapList[i].IsActiveBitTrue = false
			}
		}
	}
	return nil
}

func GetFailbackStatus(VesAuthToken string, vmmapList []draasv1alpha1.TriggerFailbackVmMapping) (bool, error) {

	bIsFailbackCompleted := true

	for i, vmmap := range vmmapList {

		//Assume Vm to be Powered off because of follow_seq
		vmmapList[i].TriggerPowerOff = true

		for j, vmdkmap := range vmmap.VmdkStatusList {

			fmt.Println("\t GetFailbackStatus: vmdkmap.SourceVmdkID :", vmdkmap.SourceVmdkID)
			fmt.Println("\t GetFailbackStatus: vmdkmap.TargetVmdkID :", vmdkmap.TargetVmdkID)

			if (vmdkmap.SourceVmdkID == "") || (vmdkmap.TargetVmdkID == "") {
				bIsFailbackCompleted = false
				fmt.Println("Continuing")
				continue
			}
			//FailbackId string
			FailbackId := vmdkmap.FailbackTriggerID
			url2 := SnifPhpUrl + "/api/failovers/"

			url2 += FailbackId

			req2, _ := http.NewRequest("GET", url2, nil)
			req2.Header.Add("content-type", "application/json")
			req2.Header.Add("cache-control", "no-cache")
			req2.Header.Add("X-VES-Authorization", VesAuthToken)

			//fmt.Println("Failback status url: ", url2)
			//fmt.Println("Request PHP API", req2)

			//skip ssl tls verify
			//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

			res2, err2 := http.DefaultClient.Do(req2)
			if err2 != nil {
				fmt.Println(err2)
			} else {
				defer res2.Body.Close()
				body2, _ := ioutil.ReadAll(res2.Body)
				var result draasv1alpha1.FailoverResponse
				if err := json.Unmarshal(body2, &result); err != nil { // Parse []byte to the go struct pointer
					fmt.Println(err)
					fmt.Println("Can not unmarshal JSON")
					bIsFailbackCompleted = false
					return false, nil
				}
				fmt.Println("GET Failback API: Failback Id : ", result.Data.Id)
				//vmdkmap.FailoverTriggerID = result.Data.Id
				fmt.Println("Failback API: Failback ACk : ", result.Data.Ack)
				vmmapList[i].VmdkStatusList[j].Ack = result.Data.Ack
				fmt.Println("GET Failback API: Failback Active flag : ", result.Data.Active)
				vmmapList[i].VmdkStatusList[j].ActiveFailback = result.Data.Active
				fmt.Println("GET Failback API: Failback Sent Blocks : ", result.Data.Sentblocks)
				vmmapList[i].VmdkStatusList[j].SentCT = result.Data.Sent_ct
				vmmapList[i].VmdkStatusList[j].SentBlocks = result.Data.Sentblocks
				vmmapList[i].VmdkStatusList[j].TotalBlocks = result.Data.TotalBlocks
				vmmapList[i].VmdkStatusList[j].Follow_Seq = result.Data.Follow_Seq
				fmt.Println("GET Failback API: Failback Total Blocks : ", result.Data.TotalBlocks)

				//TODO: Check whether its null or empty
				if vmmapList[i].VmdkStatusList[j].Follow_Seq == "" {
					fmt.Println("\n\t Follow Sequence is EMPTY.. Setting TriggerPowerOFF to False")
					vmmapList[i].TriggerPowerOff = false
				} else {
					fmt.Println("\n\t Follow Sequence is **NOT EMPTY**** vmmapList[i].TriggerPowerOff: ", vmmapList[i].TriggerPowerOff)
				}

				if vmdkmap.ActiveFailback == false {
					fmt.Println("\t Acive bit is false.. ")
					vmmapList[i].VmdkStatusList[j].RehydrationStatus = RECOVERY_ACTIVITY_COMPLETED
				} else {
					fmt.Println("\t Acive bit is true.. ")
					vmmapList[i].VmdkStatusList[j].RehydrationStatus = RECOVERY_ACTIVITY_IN_PROGRESS
					bIsFailbackCompleted = false
				}
			}
		}
	}
	return bIsFailbackCompleted, nil
}

func GetFailoverStatus(VesAuthToken string, vmmapList []draasv1alpha1.TriggerFailoverVmMapping) (bool, error) {

	bIsFailoverCompleted := true
	for i, vmmap := range vmmapList {

		for j, vmdkmap := range vmmap.VmdkStatusList {

			fmt.Println("\t GetFailoverStatus: vmdkmap.SourceVmdkID :", vmdkmap.SourceVmdkID)
			fmt.Println("\t GetFailoverStatus: vmdkmap.TargetVmdkID :", vmdkmap.TargetVmdkID)

			if (vmdkmap.SourceVmdkID == "") || (vmdkmap.TargetVmdkID == "") {
				bIsFailoverCompleted = false
				fmt.Println("Continuing")
				continue
			}
			//FailoverId string
			FailoverId := vmdkmap.FailoverTriggerID
			//vesauth, _ := ctx.Request.Cookie("VESauth")
			url2 := SnifPhpUrl + "/api/failovers/"

			url2 += FailoverId
			//var jsonStr = []byte(`{"vmdk_id":"56", "new_vmdk_id":"77"}`)
			//jsonData := map[string]string{"vmdk_id": vmdkmap.SourceVmdkID, "new_vmdk_id": vmdkmap.TargetVmdkID}
			//jsonStr, _ := json.Marshal(jsonData)

			req2, _ := http.NewRequest("GET", url2, nil)
			req2.Header.Add("content-type", "application/json")
			req2.Header.Add("cache-control", "no-cache")
			req2.Header.Add("X-VES-Authorization", VesAuthToken)

			//fmt.Println("Failover status url: ", url2)
			//fmt.Println("Request PHP API", req2)

			//skip ssl tls verify
			//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

			res2, err2 := http.DefaultClient.Do(req2)
			if err2 != nil {
				fmt.Println(err2)
			} else {
				defer res2.Body.Close()
				body2, _ := ioutil.ReadAll(res2.Body)
				var result draasv1alpha1.FailoverResponse
				if err := json.Unmarshal(body2, &result); err != nil { // Parse []byte to the go struct pointer
					fmt.Println(err)
					fmt.Println("Can not unmarshal JSON")
					bIsFailoverCompleted = false
					return false, nil
				}
				fmt.Println("GET Failover API: Failover Id : ", result.Data.Id)
				//vmdkmap.FailoverTriggerID = result.Data.Id
				fmt.Println("Failover API: Failover ACk : ", result.Data.Ack)
				vmmapList[i].VmdkStatusList[j].Ack = result.Data.Ack
				fmt.Println("GET Failover API: Failover Active flag : ", result.Data.Active)
				vmmapList[i].VmdkStatusList[j].ActiveFailover = result.Data.Active
				fmt.Println("GET Failover API: Failover Sent Blocks : ", result.Data.Sentblocks)
				vmmapList[i].VmdkStatusList[j].SentCT = result.Data.Sent_ct
				vmmapList[i].VmdkStatusList[j].SentBlocks = result.Data.Sentblocks
				vmmapList[i].VmdkStatusList[j].TotalBlocks = result.Data.TotalBlocks
				fmt.Println("GET Failover API: Failover Total Blocks : ", result.Data.TotalBlocks)

				if vmdkmap.ActiveFailover == false {
					if (vmdkmap.SentBlocks != "0") && (vmdkmap.SentBlocks == vmdkmap.SentCT) {
						vmmapList[i].VmdkStatusList[j].RehydrationStatus = RECOVERY_ACTIVITY_COMPLETED
					} else {
						vmmapList[i].VmdkStatusList[j].RehydrationStatus = RECOVERY_ACTIVITY_IN_PROGRESS
						bIsFailoverCompleted = false
					}
				} else {
					vmmapList[i].VmdkStatusList[j].RehydrationStatus = RECOVERY_ACTIVITY_IN_PROGRESS
					bIsFailoverCompleted = false
				}
			}
		}
	}
	return bIsFailoverCompleted, nil

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
	var controllerDevice types.BaseVirtualDevice
	var controller types.BaseVirtualController
	var err error

	fmt.Println("len(vmInfo.Vmdks): ", len(vmInfo.Disks))
	for i, disk := range vmInfo.Disks {
		if (i % MAXDISKSPERCONTROLLER) == 0 {
			fmt.Println("Creating new controller")
			controllerDevice, err = CreateController(&devices, "")
			if err != nil {
				fmt.Println("Failed to create controller")
				return nil, err
			}

			devices = append(devices, controllerDevice)
			controller, err = devices.FindDiskController(devices.Name(controllerDevice))
			if err != nil {
				return nil, err
			}
		}

		thin := disk.ThinProvisioned
		vDisk := &types.VirtualDisk{
			VirtualDevice: types.VirtualDevice{
				Key:        devices.NewKey(),
				UnitNumber: &controller.GetVirtualController().Key,
				Backing: &types.VirtualDiskFlatVer2BackingInfo{
					DiskMode:        string(types.VirtualDiskModePersistent),
					ThinProvisioned: &thin,
				},
			},
			CapacityInKB: disk.SizeMB * 1024,
			//TODO Iofilter: ,
		}

		devices.AssignController(vDisk, controller)
		devices = append(devices, vDisk)
	}

	return devices, nil
}

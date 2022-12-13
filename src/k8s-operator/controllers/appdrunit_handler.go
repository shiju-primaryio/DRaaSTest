package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	draasv1alpha1 "github.com/CacheboxInc/DRaaS/src/k8s-operator/api/v1alpha1"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/pbm"
	pbmtypes "github.com/vmware/govmomi/pbm/types"
	"github.com/vmware/govmomi/vim25/types"
)

func ChangePolicyState(vcenter draasv1alpha1.VCenterSpec, vmUuid string, policyAttach bool) (draasv1alpha1.VMStatus, error) {
	var VmDetails draasv1alpha1.VMStatus

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

	//Verify policy attached/detached to VM
	vmList, err := getVmList(vcenter, []string{vmUuid})
	if err != nil {
		fmt.Println("Failed to fetch VM list", err)
		return VmDetails, err
	}

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

	vmList, err = getVmList(vcenter, []string{vmUuid})
	if err != nil {
		fmt.Println("Failed to fetch VM list", err)
		return VmDetails, err
	}

	for _, vm := range vmList {
		if vm.VmUuid == vmUuid {
			fmt.Println("adding vmstatus....")
			VmDetails = vm
		}
	}

	fmt.Println("Storage policy state changed successfully to VM : ", vmObj.Name())
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

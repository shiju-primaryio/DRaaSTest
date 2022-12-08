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

func AttachPolicy(vmPolicy draasv1alpha1.VmPolicySchema) (draasv1alpha1.VmPolicyStatus, error) {
	var VmPolicyStatus draasv1alpha1.VmPolicyStatus
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := GetVCenterClient(draasv1alpha1.
		VCenterSpec{IP: vmPolicy.VcHostIp, UserName: vmPolicy.VcUsername, Password: vmPolicy.VcPassword})
	if err != nil {
		fmt.Println("Error connecting to vCenter : ", err)
		return VmPolicyStatus, err
	}

	//Get Policy details
	policyCheck := draasv1alpha1.IsPolicyExistsSchema{
		PolicyName: vmPolicy.PolicyName, VcHostIp: vmPolicy.VcHostIp,
		VcUsername: vmPolicy.VcUsername, VcPassword: vmPolicy.VcPassword}
	policyDetails, err := GetPolicy(policyCheck)
	if err != nil {
		fmt.Println("Unable to fetch policies from vCenter.")
		return VmPolicyStatus, err
	} else if policyDetails.PolicyId == "" {
		fmt.Println("Policy with given name not availabe at vCenter.")
		err = errors.New("policy with given name not availabe at vCenter")
		return VmPolicyStatus, err
	}

	vmObj, err := GetVmObject(client, vmPolicy.VmUuid)
	if err != nil {
		fmt.Println("Error getting VM : ", err)
		return VmPolicyStatus, err
	}

	deviceList, err := vmObj.Device(ctx)
	for _, device := range deviceList {
		switch disk := device.(type) {
		case *types.VirtualDisk:
			spec := types.VirtualMachineConfigSpec{}
			config := &types.VirtualDeviceConfigSpec{
				Device:    disk,
				Operation: types.VirtualDeviceConfigSpecOperationEdit,
				Profile: []types.BaseVirtualMachineProfileSpec{
					&types.VirtualMachineDefinedProfileSpec{
						ProfileId: policyDetails.PolicyId,
					},
				},
			}

			spec.DeviceChange = append(spec.DeviceChange, config)

			task, err := vmObj.Reconfigure(ctx, spec)
			if err != nil {
				return VmPolicyStatus, err
			}

			err = task.Wait(ctx)
			if err != nil {
				fmt.Println("error changing disk storage policy : ", err)
				return VmPolicyStatus, err
			}
		}
	}

	VmPolicyStatus.VmUuid = vmPolicy.VmUuid
	VmPolicyStatus.PolicyAttach = true
	fmt.Println("Storage policy attached successfully to VM : ", vmObj.Name())
	return VmPolicyStatus, nil
}

func DetachPolicy(vmPolicy draasv1alpha1.VmPolicySchema) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := GetVCenterClient(draasv1alpha1.
		VCenterSpec{IP: vmPolicy.VcHostIp, UserName: vmPolicy.VcUsername, Password: vmPolicy.VcPassword})
	if err != nil {
		fmt.Println("Error connecting to vCenter : ", err)
		return err
	}

	vmObj, err := GetVmObject(client, vmPolicy.VmUuid)
	if err != nil {
		fmt.Println("Error getting VM : ", err)
		return err
	}

	deviceList, err := vmObj.Device(ctx)
	for _, device := range deviceList {
		switch disk := device.(type) {
		case *types.VirtualDisk:
			spec := types.VirtualMachineConfigSpec{}

			config := &types.VirtualDeviceConfigSpec{
				Device:    disk,
				Operation: types.VirtualDeviceConfigSpecOperationEdit,
				Profile:   []types.BaseVirtualMachineProfileSpec{&types.VirtualMachineEmptyProfileSpec{}},
			}

			spec.DeviceChange = append(spec.DeviceChange, config)

			task, err := vmObj.Reconfigure(ctx, spec)
			if err != nil {
				return err
			}

			err = task.Wait(ctx)
			if err != nil {
				fmt.Println("error removing disk policy: ", err)
				return err
			}
		}
	}

	fmt.Println("Storage policy detached successfully from VM : ", vmObj.Name())
	return nil
}

func GetPolicy(policyCheck draasv1alpha1.IsPolicyExistsSchema) (draasv1alpha1.PolicyDetails, error) {
	var response_body draasv1alpha1.PolicyDetails
	vcenter := draasv1alpha1.VCenterSpec{
		IP: policyCheck.VcHostIp, UserName: policyCheck.VcUsername, Password: policyCheck.VcPassword,
	}

	policyList, err := GetPolicyList(vcenter)
	if err != nil {
		fmt.Println("Error fetching policy list from vCenter: ", err)
		return response_body, err
	}

	for _, policy := range policyList {
		if policy.PolicyName == policyCheck.PolicyName {
			fmt.Println("Policy available in vCenter.")
			response_body.PolicyName = policy.PolicyName
			response_body.PolicyId = policy.PolicyId
			break
		}
	}

	return response_body, err
}

//IsPolicyExists should return true/false only

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

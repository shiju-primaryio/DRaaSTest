package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	draasv1alpha1 "github.com/CacheboxInc/DRaaS/src/k8s-operator/api/v1alpha1"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/pbm"
	pbmtypes "github.com/vmware/govmomi/pbm/types"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const MAXDISKSPERCONTROLLER = 15

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

func GetAbsPath(dss []mo.Datastore, fileName string) string {
	var absolutePath string
	datstore := strings.Split(fileName, " ")[0][1:]
	dbs := datstore[:len(datstore)-1]
	vmdkPath := strings.Split(fileName, " ")[1]
	for _, ds := range dss {
		if dbs == ds.Summary.Name {
			absolutePath = (strings.Split(ds.Summary.Url, ":")[1] + vmdkPath)[2:]
		}
	}

	return absolutePath
}

func GetProfileId(vCenterClient *vim25.Client) (string, error) {
	PolicyName := "PrimaryIO_replication"
	ctx := context.Background()

	pbmSi, err := pbm.NewClient(ctx, vCenterClient)
	if err != nil {
		fmt.Println("Error creating pbm client : ", err)
		return "", err
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
		if profile.Name == PolicyName {
			break
		}
	}

	profileId := profile.ProfileId.UniqueId
	if profileId == "" {
		err = errors.New("policy not available")
		return "", err

	}

	return profile.ProfileId.UniqueId, nil
}

func CreateController(devices *object.VirtualDeviceList, controllerType string) (types.BaseVirtualDevice, error) {
	var controller types.BaseVirtualDevice
	if controllerType != "ide" {
		if controllerType == "nvme" {
			nvme, err := devices.CreateNVMEController()
			if err != nil {
				return nil, err
			}

			controller = nvme
		} else {
			scsi, err := devices.CreateSCSIController("")
			if err != nil {
				return nil, err
			}

			controller = scsi
		}
	}

	// If controller is specified to be IDE or if an ISO is specified, add IDE controller.
	if controllerType == "ide" {
		ide, err := devices.CreateIDEController()
		if err != nil {
			return nil, err
		}

		controller = ide
	}

	return controller, nil
}

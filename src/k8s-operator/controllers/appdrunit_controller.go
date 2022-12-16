/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	draasv1alpha1 "github.com/CacheboxInc/DRaaS/src/k8s-operator/api/v1alpha1"
)

// AppDRUnitReconciler reconciles a AppDRUnit object
type AppDRUnitReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=draas.primaryio.com,resources=appdrunits,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=draas.primaryio.com,resources=appdrunits/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=draas.primaryio.com,resources=appdrunits/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AppDRUnit object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *AppDRUnitReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	reqLogger := logger.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling appdr_unit Config")

	var err error

	// Fetch the AppDRUnit instance
	instance := &draasv1alpha1.AppDRUnit{}
	err = r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	fmt.Println("\n Printing info..Name: ", instance.Name)
	fmt.Println("\n Printing info..Req.NamespaceName: ", req.Namespace)
	fmt.Println("\n Printing info..req.Name: ", req.Name)
	fmt.Println("\n Printing info..Req.NamespacedName: ", req.NamespacedName)

	defer func() {
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}
	}()

	// If Site is deleted, initiate cleanup.
	if instance.DeletionTimestamp != nil {
		reqLogger.Info("Cleanup initiated for ", "Site", instance.Name)
		// TODO: Add cleanup of resources owned by Site. Site owns Storage policy created at the time of site addition
		reqLogger.Info("Cleanup successful", "Site", instance.Name)
		instance.Finalizers = nil
		if err = r.Client.Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update", "Site", instance.Name)
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	//rkReq reconsile. .Request.NamespacedName

	var ns1 types.NamespacedName //:= {Name: "site-dr-sample", Namespace: "k8s-operator-system"}
	ns1.Name = "site-dr-sample"
	ns1.Namespace = "k8s-operator-system"
	// NamespacedName{Name: "site-dr-sample", Namespace: "k8s-operator-system"}
	//rkNameSpace  req.ctrl.reconcile. NamespacedName := {"k8s-operator-system","site-dr-sample"};
	//ctrl.Request remoteSiteRequestctx := {};

	//RemoteSite_NamedSpacedName := "k8s-operator-system/site-dr-sample"
	//Fetch Site Object
	instance_dr := &draasv1alpha1.Site{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: "site-dr-sample", Namespace: "k8s-operator-system"}, instance_dr)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	var vcenter_dr draasv1alpha1.VCenterSpec
	vcenter_dr = instance_dr.Spec.VCenter
	fmt.Println("remote vcenter: ", vcenter_dr.IP)
	fmt.Println("remote username : ", vcenter_dr.UserName)

	VesAuth := instance.Spec.VesToken

	//instance.Spec.VCenter

	if instance.Spec.TriggerFailover {
		fmt.Println("Failover Step1 ")

		//Step 1: Create VM Infrastructure, make sure VM is Powered Off
		instance.Status.FailoverStatus = "Failover_Infrastructure_Creation_Started"
		for _, vm := range instance.Status.ProtectedVmList {
			fmt.Println("Creating VM : ", vm.Name)
			CreateVM(vcenter_dr, vm)
		}
		fmt.Println("Failover Step2: Listing Source VMDKs Info.. ")
		//createVMInfrastructure()
		instance.Status.FailoverStatus = "Failover_Infrastructure_Creation_Completed"
		//Step 2: Make all VMs powered OFF

		//Step 3: Attach policy if not already attached
		//ChangePolicyState(vcenter_dr, [])
		//Step 4: Fetch Information of VMs from FailedOver vCenter

		//Step 5: Create  Structure for Trigger Failover
		var vmdkmapList []draasv1alpha1.TriggerFailoverVmdkMapping

		//Step 6: add source vmdk ids from scope
		for _, vm := range instance.Status.ProtectedVmList {
			var vmdkmap draasv1alpha1.TriggerFailoverVmdkMapping
			vmdkmap.VmName = vm.Name
			vmdkmap.SourceVmUUID = vm.VmUuid
			for _, dev := range vm.Disks {
				vmdkmap.UnitNumber = int(dev.UnitNumber)
				vmdkmap.SourceScope = dev.AbsPath
				vmdkmap.Label = dev.Label

				fmt.Println("PVmUUID: ", vm.VmUuid)
				fmt.Println("PvmName : ", vm.Name)
				fmt.Println("Source Unit number : ", vmdkmap.UnitNumber)
				fmt.Println("Source scope : ", vmdkmap.SourceScope)

				vmdkmapList = append(vmdkmapList, vmdkmap)
			}
		}
		fmt.Println("Failover Step3: Listing Target VMDKs Info.. ")

		//var vmPolicyRequestList []draasv1alpha1.VmPolicyRequest
		//var vmPolicyRequest draasv1alpha1.VmPolicyRequest

		//Add TargetVMDKUUID and Scope
		// Fetch VMs from VCenter on Site Creation only
		vmList, err := getVmList(instance_dr.Spec.VCenter, nil)
		if err != nil {
			reqLogger.Error(err, "Failed to fetch VM list")
		}

		//		for _, vm := range instance_dr.Status.VmList {
		for _, vm := range vmList {
			for i, vmdkmap := range vmdkmapList {
				vmdkmap.TargetVmUUID = vm.VmUuid

				if vmdkmap.VmName == vm.Name {

					//vmPolicyRequest.VmUuid = vm.VmUuid
					//vmPolicyRequest.IsPolicyAttach = true
					//vmPolicyRequestList = append(vmPolicyRequestList, vmPolicyRequest)

					fmt.Println("Target VmUUID: ", vm.VmUuid)
					fmt.Println("Target vmName : ", vm.Name)
					vmdkmap.TargetScope = vm.VmUuid

					for _, dev := range vm.Disks {
						fmt.Println("listing Unit number : ", dev.UnitNumber)
						//if vmdkmap.UnitNumber == int(dev.UnitNumber) {
						if vmdkmap.Label == dev.Label {
							//fmt.Println("Target Unit number : ", vmdkmap.UnitNumber)

							vmdkmap.TargetScope = dev.AbsPath
							vmdkmapList[i] = vmdkmap
							fmt.Println("Target scope : ", vmdkmap.TargetScope)
						}
					}
				}
			}
		}

		/*
			fmt.Println("Failover Step4: Calling attach policy ")
			for _, vminfo := range vmPolicyRequestList {
				fmt.Println("Failover: id  ", vminfo.VmUuid)
			}
			// Attach Policy
			//ChangePolicyState(vcenter_dr, vmPolicyRequestList)
			fmt.Println("Failover Step5 ")
		*/
		fmt.Println("Failover Step4: Calling  GetVMDKsFromPostGresDB..")

		VMDKListFromPostGresDResponse, err := GetVMDKsFromPostGresDB(VesAuth, vmdkmapList)
		for _, vminfo := range VMDKListFromPostGresDResponse.Data {
			for i, vmdkmap := range vmdkmapList {
				fmt.Println("Failover : SourceScope:", vmdkmap.SourceScope)
				fmt.Println("Failover : TargetScope:", vmdkmap.TargetScope)
				fmt.Println("Failover : vminfo.VmdkScope :", vminfo.VmdkScope)

				if vmdkmap.SourceScope == vminfo.VmdkScope {
					fmt.Println("PVmscope: ", vminfo.VmdkScope)
					fmt.Println("PvmName : ", vminfo.VmdkId)
					vmdkmap.SourceVmdkID = vminfo.VmdkId
					fmt.Println("PvmvmdkId : ", vminfo.VmdkId)
					vmdkmapList[i] = vmdkmap
				}
				if vmdkmap.TargetScope == vminfo.VmdkScope {
					fmt.Println("TVmscope: ", vminfo.VmdkScope)
					fmt.Println("TvmName : ", vminfo.VmdkId)
					vmdkmap.TargetVmdkID = vminfo.VmdkId
					fmt.Println("TvmvmdkId : ", vminfo.VmdkId)

					vmdkmapList[i] = vmdkmap
				}
			}

		}
		fmt.Println("Failover Step5: Triggering failover ")

		//Step 6: PowerOff all VMs
		for _, vmdkmap := range vmdkmapList {
			VmPowerChange(vcenter_dr, vmdkmap.TargetVmUUID, false)
		}

		//Step 7: Trigger Failover
		InitiateFailover(VesAuth, vmdkmapList)
		fmt.Println("Failover Step6: Powering on VM ")

		//Step 8: PowerOn all VMs
		for _, vmdkmap := range vmdkmapList {
			VmPowerChange(vcenter_dr, vmdkmap.TargetVmUUID, true)
		}

		instance.Spec.TriggerFailover = false
		//update Spec
		if err = r.Client.Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update", "Site", instance.Name)
			return reconcile.Result{}, err
		}

	}

	instance.Status.FailoverStatus = "Failover_Not_Started"
	instance.Status.Site = "test"
	instance.Status.RemoteSite = "test-dr"

	if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
		reqLogger.Error(err, "Failed to update Site status")
	}

	var vmDetailList []draasv1alpha1.VMStatus
	if instance.Spec.ProtectVMUUIDList != nil {
		//Case 1: Protect all VM's
		if instance.Status.ProtectedVmList == nil {
			vmDetailList, err = ChangePolicyState(instance.Spec.VCenter, instance.Spec.ProtectVMUUIDList)
			if err != nil {
				fmt.Println("Failed to attach VM .......", err)
			}
		} else {
			var modifyProtectVmReq []draasv1alpha1.VmPolicyRequest

			//Case 2: Protect-Unprotect few Vm's
			var bAlreadyProtectedVMFound bool
			bAlreadyProtectedVMFound = false
			if instance.Spec.ProtectVMUUIDList != nil {
				for _, statusVM := range instance.Status.ProtectedVmList {
					for _, specVM := range instance.Spec.ProtectVMUUIDList {
						if statusVM.VmUuid == specVM.VmUuid {
							bAlreadyProtectedVMFound = true
							break
						}
					}
					//VM is already protected, but user wants to unprotect it
					if !bAlreadyProtectedVMFound {
						modifyProtectVmReq = append(modifyProtectVmReq, draasv1alpha1.VmPolicyRequest{VmUuid: statusVM.VmUuid, IsPolicyAttach: false})
						//unProtectVmUuidlist = append(unProtectVmUuidlist, specVM.VmUuid)
					}
				}

				for _, specVM := range instance.Spec.ProtectVMUUIDList {
					for _, statusVM := range instance.Status.ProtectedVmList {
						if statusVM.VmUuid == specVM.VmUuid {
							bAlreadyProtectedVMFound = true
							break
						}
					}
					//VM is not protected, but user wants to protect it
					if !bAlreadyProtectedVMFound {
						modifyProtectVmReq = append(modifyProtectVmReq, specVM)
					}
				}

			} else {
				//Case 3: Unprotect All VM's
				for _, specVM := range instance.Spec.ProtectVMUUIDList {
					modifyProtectVmReq = append(modifyProtectVmReq, draasv1alpha1.VmPolicyRequest{VmUuid: specVM.VmUuid, IsPolicyAttach: false})
				}
			}

			vmDetailList, err = ChangePolicyState(instance.Spec.VCenter, modifyProtectVmReq)
		}

		/*
			for _, vm := range vmDetailList {
				fmt.Println("VmUUID: ", vm.VmUuid)
				fmt.Println("vmName : ", vm.Name)
			}
		*/
		instance.Status.ProtectedVmList = vmDetailList
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}

	}

	fmt.Println("Setting ProtectVMUUIDList to Nil .......")
	instance.Spec.ProtectVMUUIDList = nil
	//update Spec
	if err = r.Client.Update(context.TODO(), instance); err != nil {
		reqLogger.Error(err, "Failed to update", "Site", instance.Name)
		return reconcile.Result{}, err
	}

	for _, vm := range instance.Status.ProtectedVmList {
		fmt.Println("PVmUUID: ", vm.VmUuid)
		fmt.Println("PvmName : ", vm.Name)
	}

	if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
		reqLogger.Error(err, "Failed to update Site status")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppDRUnitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&draasv1alpha1.AppDRUnit{}).
		Complete(r)
}

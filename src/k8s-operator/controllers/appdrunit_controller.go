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
	"time"

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

const RECOVERY_ACTIVITY_NOT_STARTED string = "NOT_STARTED"
const RECOVERY_ACTIVITY_STARTED string = "STARTED"
const RECOVERY_ACTIVITY_IN_PROGRESS string = "IN_PROGRESS"
const RECOVERY_ACTIVITY_COMPLETED string = "COMPLETED"
const RECOVERY_ACTIVITY_ERROR_OCCURED string = "ERROR_OCCURED"

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

	//fmt.Println("\n **AppDRUnit** Current date and time is: ", time.Now().String())

	// TODO(user): your logic here
	reqLogger := logger.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling appdr_unit Config")

	//fmt.Println("\n AppDRUnit Current date and time is: ", time.Now().String())
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

	/*
		fmt.Println("\n Printing info..Name: ", instance.Name)
		fmt.Println("\n Printing info..Req.NamespaceName: ", req.Namespace)
		fmt.Println("\n Printing info..req.Name: ", req.Name)
		fmt.Println("\n Printing info..Req.NamespacedName: ", req.NamespacedName)
	*/

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

	var vcenter_dr draasv1alpha1.VCenterSpec
	var instance_dr *draasv1alpha1.Site

	//instance.Spec.Site = "site-sample"
	//instance.Spec.PeerSite = "site-dr-sample"
	if instance.Spec.Site != "" {
		instance.Status.Site = instance.Spec.Site
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}
	}

	if instance.Spec.PeerSite != "" {
		instance.Status.PeerSite = instance.Spec.PeerSite
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}

		//Fetch Peer Site Object
		instance_dr = &draasv1alpha1.Site{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Status.PeerSite, Namespace: "k8s-operator-system"}, instance_dr)
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
		vcenter_dr = instance_dr.Spec.VCenter
		fmt.Println("remote vcenter: ", vcenter_dr.IP)
		fmt.Println("remote username : ", vcenter_dr.UserName)
	}

	if (instance.Spec.TriggerFailover) && (instance_dr != nil) && (instance.Status.FailoverStatus.InfrastructureStatus != RECOVERY_ACTIVITY_COMPLETED) {
		fmt.Println("Failover Step1 ")

		//Update Status
		instance.Status.FailoverStatus.OverallFailoverStatus = RECOVERY_ACTIVITY_STARTED
		instance.Status.FailoverStatus.InfrastructureStatus = RECOVERY_ACTIVITY_STARTED
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}

		//Step 1: Create VM Infrastructure, make sure VM is Powered Off
		for _, vm := range instance.Status.ProtectedVmList {
			fmt.Println("Creating VM : ", vm.Name)
			CreateVM(vcenter_dr, vm)
		}

		//Update Status
		instance.Status.FailoverStatus.InfrastructureStatus = RECOVERY_ACTIVITY_COMPLETED
		instance.Status.FailoverStatus.OverallFailoverStatus = RECOVERY_ACTIVITY_IN_PROGRESS
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}

		//Optional step: Attach policy if not already attached
		//ChangePolicyState(vcenter_dr, [])

		fmt.Println("Failover Step2: Listing Source VMDKs Info.. ")

		var vmdkmapList []draasv1alpha1.TriggerFailoverVmdkMapping

		//Step 2: Add source vmdk ids from scope
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

		//Add TargetVMDKUUID and Scope
		// Fetch VMs from VCenter on Site Creation only
		vmList, err := getVmList(instance_dr.Spec.VCenter, nil)
		if err != nil {
			reqLogger.Error(err, "Failed to fetch VM list")
		}

		for _, vm := range vmList {
			for i, vmdkmap := range vmdkmapList {
				vmdkmap.TargetVmUUID = vm.VmUuid

				if vmdkmap.VmName == vm.Name {

					fmt.Println("Target VmUUID: ", vm.VmUuid)
					fmt.Println("Target vmName : ", vm.Name)
					vmdkmap.TargetScope = vm.VmUuid

					for _, dev := range vm.Disks {
						//Compare Label of VMDK
						if vmdkmap.Label == dev.Label {

							vmdkmap.TargetScope = dev.AbsPath
							vmdkmapList[i] = vmdkmap
							fmt.Println("Target scope : ", vmdkmap.TargetScope)
						}
					}
				}
			}
		}
		instance.Status.FailoverVmdkListStatus = vmdkmapList

		if err = r.Client.Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update", "Site", instance.Name)
			return reconcile.Result{}, err
		}

		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}

		fmt.Println("Failover Step4: Calling  GetVMDKsFromPostGresDB to create map..")

		VMDKListFromPostGresDResponse, err := GetVMDKsFromPostGresDB(instance.Spec.VesToken, instance.Spec.SnifPhpUrl)
		for _, vminfo := range VMDKListFromPostGresDResponse.Data {
			for i, vmdkmap := range vmdkmapList {
				fmt.Println("Failover : SourceScope:", vmdkmap.SourceScope)
				fmt.Println("Failover : TargetScope:", vmdkmap.TargetScope)
				fmt.Println("Failover : vminfo.VmdkScope :", vminfo.VmdkScope)

				if len(vmdkmap.SentBlocks) == 0 {
					vmdkmap.SentCT = "0"
					vmdkmap.SentBlocks = "0"
					vmdkmap.TotalBlocks = "0"
				}
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

		//Step 6: PowerOff all VMs

		//Update Status : activity started
		fmt.Println("Failover Step5: Powering off VMs.. ")
		instance.Status.FailoverStatus.PowerOffStatus = RECOVERY_ACTIVITY_STARTED
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}

		for _, vmdkmap := range vmdkmapList {
			VmPowerChange(vcenter_dr, vmdkmap.TargetVmUUID, false)
		}

		//Update Status : activity Completed
		instance.Status.FailoverStatus.PowerOffStatus = RECOVERY_ACTIVITY_COMPLETED
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}

		//Step 7: Trigger Failover
		fmt.Println("Failover Step5: Triggering failover ")
		instance.Status.FailoverVmdkListStatus = vmdkmapList
		instance.Status.FailoverStatus.RehydrationStatus = RECOVERY_ACTIVITY_STARTED
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}
	}

	if (instance.Spec.TriggerFailover) && (instance_dr != nil) {
		var vmdkmapList []draasv1alpha1.TriggerFailoverVmdkMapping

		vmdkmapList = instance.Status.FailoverVmdkListStatus

		InitiateFailover(instance.Spec.VesToken, instance.Spec.SnifPhpUrl, vmdkmapList)

		//Update Status : activity Completed
		instance.Status.FailoverStatus.PowerOffStatus = RECOVERY_ACTIVITY_IN_PROGRESS
		instance.Status.FailoverVmdkListStatus = vmdkmapList
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}

		fmt.Println("Sleep Over for 10 seconds for active bit to be set.....")
		// Calling Sleep method
		time.Sleep(10 * time.Second)

		fmt.Println("Failover Step6: Waiting for Active bit to be set for all protected VMs ")
		WaitForActiveBitTobeSet(instance.Spec.VesToken, instance.Spec.SnifPhpUrl, instance.Status.FailoverVmdkListStatus)

		fmt.Println("Failover Step6: Powering on VM ")

		//Step 8: PowerOn all VMs
		instance.Status.FailoverStatus.PowerOnStatus = RECOVERY_ACTIVITY_STARTED
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}
		// Calling Sleep method
		fmt.Println("Sleep Over for 10 seconds before powering ON.....")
		time.Sleep(10 * time.Second)

		bAllVmDKsActive := true
		for _, vmdkmap := range vmdkmapList {
			if vmdkmap.ActiveFailover {
				VmPowerChange(vcenter_dr, vmdkmap.TargetVmUUID, true)
			} else {
				instance.Status.FailoverStatus.PowerOnStatus = RECOVERY_ACTIVITY_IN_PROGRESS
				bAllVmDKsActive = false
			}
		}

		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}

		//update Spec
		if bAllVmDKsActive {
			instance.Status.FailoverStatus.PowerOnStatus = RECOVERY_ACTIVITY_COMPLETED
			instance.Status.FailoverStatus.RehydrationStatus = RECOVERY_ACTIVITY_IN_PROGRESS
			instance.Spec.TriggerFailover = false
		}
		if err = r.Client.Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update", "Site", instance.Name)
			return reconcile.Result{}, err
		}

		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}

		fmt.Println("Sleep Over for 10 seconds after failover trigger.....")
		// Calling Sleep method
		time.Sleep(10 * time.Second)

		//Requeue the work to reconsiler
		if err != nil {
			return reconcile.Result{Requeue: true}, nil
		}
	}

	if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
		reqLogger.Error(err, "Failed to update Site status")
	}

	if (instance.Spec.TriggerCancelRecoveryOperation) && (instance_dr != nil) {
		instance.Status.FailoverStatus.OverallFailoverStatus = RECOVERY_ACTIVITY_NOT_STARTED
		instance.Status.FailoverStatus.InfrastructureStatus = RECOVERY_ACTIVITY_NOT_STARTED
		instance.Status.FailoverStatus.RehydrationStatus = RECOVERY_ACTIVITY_NOT_STARTED
		instance.Status.FailoverStatus.PowerOnStatus = RECOVERY_ACTIVITY_NOT_STARTED
		instance.Status.FailoverStatus.PowerOffStatus = RECOVERY_ACTIVITY_NOT_STARTED

	}

	if instance.Status.FailoverStatus.OverallFailoverStatus == RECOVERY_ACTIVITY_IN_PROGRESS {
		fmt.Println("Adding GetFailoverStatus .......")

		bIsFailoverCompleted, _ := GetFailoverStatus(instance.Spec.VesToken, instance.Spec.SnifPhpUrl, instance.Status.FailoverVmdkListStatus)
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}
		if bIsFailoverCompleted {
			instance.Status.FailoverStatus.OverallFailoverStatus = RECOVERY_ACTIVITY_COMPLETED
			instance.Status.FailoverStatus.RehydrationStatus = RECOVERY_ACTIVITY_COMPLETED

			var vmdkmapList []draasv1alpha1.TriggerFailoverVmdkMapping

			vmdkmapList = instance.Status.FailoverVmdkListStatus
			fmt.Println("Failover COMPLETED.. Powering off VMs on traget.......")

			for _, vmdkmap := range vmdkmapList {
				VmPowerChange(vcenter_dr, vmdkmap.TargetVmUUID, false)
			}

			if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
				reqLogger.Error(err, "Failed to update Site status")
			}

		}
	}

	var vmDetailList []draasv1alpha1.VMStatus

	if (instance.Status.ProtectedVmList == nil) && (instance.Spec.ProtectVMUUIDList != nil) {
		//Case 1: Protect all VM's
		fmt.Println("Case 1:  Protect all VMs  .......")

		vmDetailList, err = ChangePolicyState(instance.Spec.VCenter, instance.Spec.ProtectVMUUIDList)
		if err != nil {
			fmt.Println("Failed to attach VM .......", err)
		}
		instance.Status.ProtectedVmList = vmDetailList
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}

		//for _, vmDet := range vmDetailList {
		//Trigger PowerOn for VM
		//fmt.Println("Powering ON VM :", vmDet.Name)
		//VmPowerChange(vcenter_dr, vmDet.VmUuid, true)
		//}

	} else {
		var modifyProtectVmReq []draasv1alpha1.VmPolicyRequest

		//Case 2: Protect-Unprotect few Vm's
		var bAlreadyProtectedVMFound bool
		if instance.Spec.ProtectVMUUIDList != nil {
			for _, statusVM := range instance.Status.ProtectedVmList {
				bAlreadyProtectedVMFound = false
				for _, specVM := range instance.Spec.ProtectVMUUIDList {
					if statusVM.VmUuid == specVM.VmUuid {
						bAlreadyProtectedVMFound = true
						break
					}
				}
				//VM is already protected, but user wants to unprotect it
				if !bAlreadyProtectedVMFound {
					fmt.Println("Case 2:  Unprotecting VM (status VM UUID) :", statusVM.VmUuid)

					modifyProtectVmReq = append(modifyProtectVmReq, draasv1alpha1.VmPolicyRequest{VmUuid: statusVM.VmUuid, IsPolicyAttach: false})
				}
			}
			for _, specVM := range instance.Spec.ProtectVMUUIDList {
				bAlreadyProtectedVMFound = false
				for _, statusVM := range instance.Status.ProtectedVmList {
					if statusVM.VmUuid == specVM.VmUuid {
						bAlreadyProtectedVMFound = true
						break
					}
				}
				//VM is not protected, but user wants to protect it
				if !bAlreadyProtectedVMFound {
					fmt.Println("Case 3:  Protecting VM (spec VM UUID) :", specVM.VmUuid)
					modifyProtectVmReq = append(modifyProtectVmReq, specVM)
				}
			}

		} else {
			//Case 4: Unprotect All VM's
			fmt.Println("Case 4:  UnProtect all VMs  .......")

			for _, specVM := range instance.Spec.ProtectVMUUIDList {
				modifyProtectVmReq = append(modifyProtectVmReq, draasv1alpha1.VmPolicyRequest{VmUuid: specVM.VmUuid, IsPolicyAttach: false})
			}
		}

		if len(modifyProtectVmReq) != 0 {
			fmt.Println("Length of ModifyVMProtect Request is: ", len(modifyProtectVmReq))

			vmDetailList, err = ChangePolicyState(instance.Spec.VCenter, modifyProtectVmReq)

			for _, vmProtectReq := range modifyProtectVmReq {
				if !vmProtectReq.IsPolicyAttach {
					for i, vmStatusVM := range instance.Status.ProtectedVmList {
						if vmStatusVM.VmUuid == vmProtectReq.VmUuid {
							instance.Status.ProtectedVmList = append(instance.Status.ProtectedVmList[:i], instance.Status.ProtectedVmList[i+1:]...)
						}
					}
				}
			}
			for _, vmDet := range vmDetailList {
				instance.Status.ProtectedVmList = append(instance.Status.ProtectedVmList, vmDet)
				//Trigger PowerOn for VM
				//fmt.Println("Powering ON VM :", vmDet.Name)
				//VmPowerChange(vcenter_dr, vmDet.VmUuid, true)
			}

			if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
				reqLogger.Error(err, "Failed to update Site status")
			}
		}

	}

	if instance.Status.ProtectedVmList != nil {
		fmt.Println("Getting Initial Sync details: Calling GetVMDKsFromPostGresDB to get received blocks/IO info.......")

		VMDKListFromPostGresDResponse, _ := GetVMDKsFromPostGresDB(instance.Spec.VesToken, instance.Spec.SnifPhpUrl)
		for _, vmdkmapinfo := range VMDKListFromPostGresDResponse.Data {
			for i, vminfo := range instance.Status.ProtectedVmList {
				bUpdatedVMDKInfo := false
				for j, vmdkinfo := range vminfo.Disks {
					//fmt.Println("Checking VM .......", vminfo.Name)
					if vmdkmapinfo.VmdkScope == vmdkinfo.AbsPath {
						fmt.Println("Adding Received blocks .......")

						vmdkinfo.ReceivedBlocks = vmdkmapinfo.ReceivedBlocks
						vmdkinfo.ReceivedIOs = vmdkmapinfo.ReceivedIOs
						vmdkinfo.TotalBlocks = vmdkmapinfo.TotalBlocks
						bUpdatedVMDKInfo = true
						vminfo.Disks[j] = vmdkinfo
						break
					}
					/*
						if bUpdatedVMDKInfo {
							vminfo.Disks[j] = vmdkinfo
							break
						}
					*/

				}
				if bUpdatedVMDKInfo {
					instance.Status.ProtectedVmList[i] = vminfo
					if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
						reqLogger.Error(err, "Failed to update Site status")
					}

					break
				}
			}
		}
		/*
			for _, vm := range instance.Status.ProtectedVmList {
				fmt.Println("PVmUUID: ", vm.VmUuid)
				fmt.Println("PvmName : ", vm.Name)
			}
		*/
		//update Spec
		if err = r.Client.Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update", "Site", instance.Name)
			return reconcile.Result{}, err
		}

		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}

		fmt.Println("Sleep Over for 10 seconds.....")
		// Calling Sleep method
		time.Sleep(10 * time.Second)

		// Calling reconsiler after some period
		fmt.Println("Calling reconsiler after 10 seconds/ testing... : ")
		return ctrl.Result{RequeueAfter: 10000}, nil

	}
	//update Spec
	if err = r.Client.Update(context.TODO(), instance); err != nil {
		reqLogger.Error(err, "Failed to update", "Site", instance.Name)
		return reconcile.Result{}, err
	}

	if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
		reqLogger.Error(err, "Failed to update Site status")
	}

	return reconcile.Result{Requeue: true}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppDRUnitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&draasv1alpha1.AppDRUnit{}).
		Complete(r)
}

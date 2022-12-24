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
	//reqLogger.Info("Reconciling appdr_unit Config")

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
			reqLogger.Error(err, "defer:1: Failed to update Appdrunit status")
		}
	}()

	// If Site is deleted, initiate cleanup.
	if instance.DeletionTimestamp != nil {
		reqLogger.Info("Cleanup initiated for ", "Site", instance.Name)
		// TODO: Add cleanup of resources owned by Site. Site owns Storage policy created at the time of site addition
		reqLogger.Info("Cleanup successful", "Site", instance.Name)
		instance.Finalizers = nil
		if err = r.Client.Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "2. Failed to update", "Appdrunit Site", instance.Name)
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
			reqLogger.Error(err, "3. Failed to update Appdrunit status")
		}
	}
	vcenter := instance.Spec.VCenter

	if instance.Spec.PeerSite != "" {
		instance.Status.PeerSite = instance.Spec.PeerSite
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "4. Failed to update Appdrunit status")
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
		//fmt.Println("remote vcenter: ", vcenter_dr.IP)
		//fmt.Println("remote username : ", vcenter_dr.UserName)
	}

	//Trigger Failback
	if (instance.Spec.TriggerFailback) || (instance.Spec.TriggerFailbackWithLiveReplication) {
		fmt.Println("Failback Step1 ")
		//Step 1: Create Infrastructure : Assume its already created for failback
		instance.Status.FailbackStatus.InfrastructureStatus = RECOVERY_ACTIVITY_COMPLETED

		//Update Status
		instance.Status.FailbackStatus.OverallFailbackStatus = RECOVERY_ACTIVITY_STARTED

		var vmmap_fb_List []draasv1alpha1.TriggerFailbackVmMapping
		var vmmap_fb draasv1alpha1.TriggerFailbackVmMapping

		//var vmmap_fb_List []draasv1alpha1.TriggerFailbackVmdkMapping

		var vmdkmap_fb draasv1alpha1.TriggerFailbackVmdkMapping
		fmt.Println("Failback Step2: Listing Source VMDKs Info.. Creating structure.. ")
		//Copy the Mapping from FailoverVmdkListStatus to FailbackVmdkListStatus
		for _, vmmap_fo := range instance.Status.FailoverVmListStatus {

			//Swap the fields
			vmmap_fb.SourceVmUUID = vmmap_fo.TargetVmUUID
			fmt.Println("FB: PVmUUID: ", vmmap_fb.SourceVmUUID)

			vmmap_fb.TargetVmUUID = vmmap_fo.SourceVmUUID
			fmt.Println("FB: TVmUUID: ", vmmap_fb.TargetVmUUID)
			vmmap_fb.VmName = vmmap_fo.VmName
			fmt.Println("FB: TVmUUID: ", vmmap_fb.VmName)
			vmmap_fb.IsActiveBitTrue = false
			var vmdkmap_fb_List []draasv1alpha1.TriggerFailbackVmdkMapping
			for _, vmdkmap_fo := range vmmap_fo.VmdkStatusList {

				//Swap the fields
				vmdkmap_fb.SourceScope = vmdkmap_fo.TargetScope
				vmdkmap_fb.TargetScope = vmdkmap_fo.SourceScope
				fmt.Println("FB: TVSourceScope: ", vmdkmap_fb.SourceScope)

				fmt.Println("FB: TVTargetScope: ", vmdkmap_fb.TargetScope)
				vmdkmap_fb.SourceVmdkID = vmdkmap_fo.TargetVmdkID
				vmdkmap_fb.TargetVmdkID = vmdkmap_fo.SourceVmdkID
				fmt.Println("FB: vmdkmap_fb.SourceVmdkID: ", vmdkmap_fb.SourceVmdkID)
				fmt.Println("FB: vmdkmap_fb.TargetVmdkID: ", vmdkmap_fb.TargetVmdkID)

				//Copy the fields
				vmdkmap_fb.Label = vmdkmap_fo.Label
				vmdkmap_fb.UnitNumber = vmdkmap_fo.UnitNumber
				//This field is not needed as failback vm is already created
				vmdkmap_fb.ScsiControllerId = vmdkmap_fo.ScsiControllerId

				// Initialize the fields
				vmdkmap_fb.FailbackTriggerID = ""
				vmdkmap_fb.Ack = "0"
				vmdkmap_fb.SentCT = "0"
				vmdkmap_fb.SentBlocks = "0"
				vmdkmap_fb.TotalBlocks = "0"
				vmdkmap_fb.RehydrationStatus = RECOVERY_ACTIVITY_NOT_STARTED
				vmdkmap_fb_List = append(vmdkmap_fb_List, vmdkmap_fb)
			}
			vmmap_fb.VmdkStatusList = vmdkmap_fb_List
			vmmap_fb_List = append(vmmap_fb_List, vmmap_fb)
		}

		instance.Status.FailbackVmListStatus = vmmap_fb_List

		//Update Status : activity started
		instance.Status.FailbackStatus.PowerOffStatus = RECOVERY_ACTIVITY_STARTED
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "5. Failed to update Appdrunit status")
		}

		//Step 2: PowerOff all Original VMs : Primary site
		//fmt.Println("Failback Step3: Powering off VMs on source/Primary Site.. ")
		//for _, vmmap := range vmmap_fb_List {
		//	VmPowerChange(vcenter, vmmap.TargetVmUUID, false)
		//}

		//Update Status : activity Completed
		instance.Status.FailbackStatus.PowerOffStatus = RECOVERY_ACTIVITY_COMPLETED
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "6. Failed to update Appdrunit status")
		}

		//Step 3: Trigger Failback
		fmt.Println("Failback Step4: Triggering failback ")
		instance.Status.FailbackVmListStatus = vmmap_fb_List
		instance.Status.FailbackStatus.RehydrationStatus = RECOVERY_ACTIVITY_STARTED
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "7. Failed to update Appdrunit status")
		}

		InitiateFailback(instance.Spec.VesToken, instance.Spec.TriggerFailbackWithLiveReplication, vmmap_fb_List)

		//Update Status : activity Completed
		//instance.Status.FailbackStatus.PowerOffStatus = RECOVERY_ACTIVITY_IN_PROGRESS
		instance.Status.FailbackVmListStatus = vmmap_fb_List
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "8. Failed to update Appdrunit status")
		}
		instance.Spec.TriggerFailback = false
		instance.Spec.TriggerFailbackWithLiveReplication = false
		if err = r.Client.Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "11. Failed to update Appdrunit", "Site", instance.Name)
			return reconcile.Result{}, err
		}

		fmt.Println("Sleep Over for 5 seconds for active bit to be set.....")

		// Calling Sleep method
		time.Sleep(5 * time.Second)
	}

	if instance.Status.FailbackStatus.OverallFailbackStatus == RECOVERY_ACTIVITY_STARTED {

		fmt.Println("Failback Step5: Waiting for Active bit to be set for all protected VMs ")
		WaitForActiveBitTobeSetFailBack(instance.Spec.VesToken, instance.Status.FailbackVmListStatus)

		instance.Status.FailbackStatus.PowerOnStatus = RECOVERY_ACTIVITY_STARTED
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "9. Failed to update Appdrunit status")
		}
		// Calling Sleep method
		fmt.Println("Sleep Over for 5 seconds before powering ON.....")
		time.Sleep(5 * time.Second)

		//Step 6: PowerOn all VMs
		bAllVmDKsActive := true
		for _, vmmap := range instance.Status.FailbackVmListStatus {
			//vmmap.IsActiveBitTrue = true
			//fmt.Println("\n\n ----After ActiveBitSet: powering ON VM.....", vmmap.TargetVmUUID)

			if vmmap.IsActiveBitTrue {
				//fmt.Println("\n\n After ActiveBitSet: powering ON VM.....", vmmap.TargetVmUUID)
				fmt.Println("Step 6: After ActiveBitSet: powering ON VMname.....", vmmap.VmName)

				errorStr, err := VmPowerChange(vcenter, vmmap.TargetVmUUID, true, false)
				if err != nil {
					fmt.Println("Error Powering ON VM with Vcenter IP: ", vcenter_dr.IP, " Error:", errorStr, " ErrorId: ", err)
					bAllVmDKsActive = false
				}

			} else {
				instance.Status.FailbackStatus.PowerOnStatus = RECOVERY_ACTIVITY_IN_PROGRESS
				bAllVmDKsActive = false
				fmt.Println("After ActiveBitSet: bAllVmDKsActive = false VMname.....", vmmap.VmName)
			}
		}

		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "10 Failed to update Appdrunit status")
		}

		//update Spec
		if bAllVmDKsActive {
			fmt.Println("bAllVmDKsActive is set.....")
			instance.Status.FailbackStatus.OverallFailbackStatus = RECOVERY_ACTIVITY_IN_PROGRESS
			instance.Status.FailbackStatus.PowerOnStatus = RECOVERY_ACTIVITY_COMPLETED
		}
		instance.Status.FailbackStatus.RehydrationStatus = RECOVERY_ACTIVITY_IN_PROGRESS

		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "12 Failed to update Appdrunit status")
		}

		fmt.Println("Sleep Over for 10 seconds after failback trigger.....")
		// Calling Sleep method
		time.Sleep(10 * time.Second)

		//Requeue the work to reconsiler
		if err != nil {
			return reconcile.Result{Requeue: true}, nil
		}

	}

	if (instance.Spec.TriggerFailover) && (instance_dr != nil) && (instance.Status.FailoverStatus.InfrastructureStatus != RECOVERY_ACTIVITY_COMPLETED) {
		fmt.Println("Failover Step1 ")

		//Update Status
		instance.Status.FailoverStatus.OverallFailoverStatus = RECOVERY_ACTIVITY_STARTED
		instance.Status.FailoverStatus.InfrastructureStatus = RECOVERY_ACTIVITY_STARTED
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "13 Failed to update Appdrunit status")
		}

		//Step 1: Create VM Infrastructure, make sure VM is Powered Off
		for _, vm := range instance.Status.ProtectedVmList {
			fmt.Println("Creating VM : ", vm.Name)
			CreateVM(vcenter_dr, vm)
		}

		//Update Status
		instance.Status.FailoverStatus.InfrastructureStatus = RECOVERY_ACTIVITY_COMPLETED
		//instance.Status.FailoverStatus.OverallFailoverStatus = RECOVERY_ACTIVITY_IN_PROGRESS
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "14. Appdrunit Failed to update Appdrunit status")
		}

		//Optional step: Attach policy if not already attached
		//ChangePolicyState(vcenter_dr, [])

		fmt.Println("Failover Step2: Listing Source VMDKs Info.. ")

		var vmmapList []draasv1alpha1.TriggerFailoverVmMapping

		//Step 2: Add source vmdk ids from scope
		for _, vm := range instance.Status.ProtectedVmList {
			var vmmap draasv1alpha1.TriggerFailoverVmMapping
			var vmdkmap draasv1alpha1.TriggerFailoverVmdkMapping
			var vmdkmapList []draasv1alpha1.TriggerFailoverVmdkMapping

			vmmap.VmName = vm.Name
			vmmap.SourceVmUUID = vm.VmUuid

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

			vmmap.VmdkStatusList = vmdkmapList
			vmmapList = append(vmmapList, vmmap)
		}

		fmt.Println("Failover Step3: Listing Target VMDKs Info.. ")

		//Add TargetVMDKUUID and Scope
		// Fetch VMs from VCenter on Site Creation only
		vmList, err := getVmList(instance_dr.Spec.VCenter, nil)
		if err != nil {
			reqLogger.Error(err, "Failed to fetch VM list")
		}

		for _, vm := range vmList {

			for i, vmmap := range vmmapList {
				//for i, vmdkmap := range vmdkmapList {

				if vmmap.VmName == vm.Name {

					vmmapList[i].TargetVmUUID = vm.VmUuid

					fmt.Println("Target VmUUID: ", vm.VmUuid)
					fmt.Println("Target vmName : ", vm.Name)
					//vmdkmap.TargetScope = vm.VmUuid
					vmdkmapList := vmmap.VmdkStatusList
					for j, vmdkmap := range vmdkmapList {
						for _, dev := range vm.Disks {
							//Compare Label of VMDK
							if vmdkmap.Label == dev.Label {
								//Update the scope
								vmmapList[i].VmdkStatusList[j].TargetScope = dev.AbsPath
								vmdkmap.TargetScope = dev.AbsPath
								fmt.Println("Target scope : ", vmdkmap.TargetScope)

								fmt.Println("Creating VMDK Entries for VM in Postgres DB: ", vm.Name)
								//Create VMDKs
								err := CreateVMDKsAtPostGresDB(instance.Spec.VesToken, vm)
								if err != nil {
									fmt.Println("Failed to create VMDK's at postgres .......", err)
								}

							}
						}
					}
				}
			}
		}
		instance.Status.FailoverVmListStatus = vmmapList

		if err = r.Client.Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "15 Failed to update Appdrunit", "Site", instance.Name)
			return reconcile.Result{}, err
		}

		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "16 Failed to update Appdrunit status")
		}

		fmt.Println("Failover Step4: Calling  GetVMDKsFromPostGresDB to create map..")

		VMDKListFromPostGresDResponse, err := GetVMDKsFromPostGresDB(instance.Spec.VesToken)
		if err != nil {
			reqLogger.Error(err, "Failed to fetch VMDK's from DB")
		}

		for _, vminfo := range VMDKListFromPostGresDResponse.Data {
			for i, _ := range vmmapList {
				for j, vmdkmap := range vmmapList[i].VmdkStatusList {
					fmt.Println("Failover : SourceScope:", vmdkmap.SourceScope)
					fmt.Println("Failover : TargetScope:", vmdkmap.TargetScope)
					fmt.Println("Failover : vminfo.VmdkScope :", vminfo.VmdkScope)

					if len(vmdkmap.SentBlocks) == 0 {
						vmmapList[i].VmdkStatusList[j].SentCT = "0"
						vmmapList[i].VmdkStatusList[j].SentBlocks = "0"
						vmmapList[i].VmdkStatusList[j].TotalBlocks = "0"
					}
					if vmdkmap.SourceScope == vminfo.VmdkScope {
						fmt.Println("PVmscope: ", vminfo.VmdkScope)
						fmt.Println("PvmName : ", vminfo.VmdkId)

						vmmapList[i].VmdkStatusList[j].SourceVmdkID = vminfo.VmdkId
						fmt.Println("PvmvmdkId : ", vminfo.VmdkId)
					}
					if vmdkmap.TargetScope == vminfo.VmdkScope {
						fmt.Println("TVmscope: ", vminfo.VmdkScope)
						fmt.Println("TvmName : ", vminfo.VmdkId)
						vmmapList[i].VmdkStatusList[j].TargetVmdkID = vminfo.VmdkId
						fmt.Println("TvmvmdkId : ", vminfo.VmdkId)
					}
				}
			}

		}

		//Step 4: PowerOff all VMs

		//Update Status : activity started
		fmt.Println("Failover Step5: Powering off VMs.. ")
		instance.Status.FailoverStatus.PowerOffStatus = RECOVERY_ACTIVITY_STARTED
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "17 Failed to update Appdrunit status")
		}

		for _, vmmap := range vmmapList {
			VmPowerChange(vcenter_dr, vmmap.TargetVmUUID, false, false)
		}

		//Update Status : activity Completed
		instance.Status.FailoverStatus.PowerOffStatus = RECOVERY_ACTIVITY_COMPLETED
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "18 Failed to update Appdrunit status")
		}

		instance.Status.FailoverVmListStatus = vmmapList
		instance.Status.FailoverStatus.RehydrationStatus = RECOVERY_ACTIVITY_STARTED
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "19 Failed to update Appdrunit status")
		}
	}

	if (instance.Spec.TriggerFailover) && (instance_dr != nil) {
		//var vmdkmapList []draasv1alpha1.TriggerFailoverVmdkMapping

		vmmapList := instance.Status.FailoverVmListStatus

		//Step 7: Trigger Failover
		fmt.Println("Failover Step5: Triggering failover ")

		InitiateFailover(instance.Spec.VesToken, vmmapList)

		//Update Status : activity Completed
		instance.Status.FailoverVmListStatus = vmmapList
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "20 Failed to update Appdrunit status")
		}
		instance.Spec.TriggerFailover = false
		if err = r.Client.Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "23. Failed to update Appdrunit", "Site", instance.Name)
			return reconcile.Result{}, err
		}
		fmt.Println("Sleep Over for 10 seconds for active bit to be set.....")
		// Calling Sleep method
		time.Sleep(10 * time.Second)
	}

	if instance.Status.FailoverStatus.OverallFailoverStatus == RECOVERY_ACTIVITY_STARTED {

		instance.Status.FailoverStatus.OverallFailoverStatus = RECOVERY_ACTIVITY_IN_PROGRESS
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "21 Failed to update Appdrunit status")
		}

		fmt.Println("Failover Step6: Waiting for Active bit to be set for all protected VMs ")
		WaitForActiveBitTobeSet(instance.Spec.VesToken, instance.Status.FailoverVmListStatus)

		fmt.Println("Failover Step6: Powering on VM ")

		//Step 8: PowerOn all VMs
		instance.Status.FailoverStatus.PowerOnStatus = RECOVERY_ACTIVITY_STARTED
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "21 Failed to update Appdrunit status")
		}
		// Calling Sleep method
		fmt.Println("Sleep Over for 10 seconds before powering ON.....")
		time.Sleep(10 * time.Second)

		bAllVmDKsActive := true
		for _, vmdkmap := range instance.Status.FailoverVmListStatus {
			if vmdkmap.IsActiveBitTrue {
				errorStr, err := VmPowerChange(vcenter_dr, vmdkmap.TargetVmUUID, true, false)
				if err != nil {
					fmt.Println("Error Powering ON VM with Vcenter IP: ", vcenter_dr.IP, " Error:", errorStr, " ErrorId: ", err)
					bAllVmDKsActive = false
				}
			} else {
				instance.Status.FailoverStatus.PowerOnStatus = RECOVERY_ACTIVITY_IN_PROGRESS
				bAllVmDKsActive = false
			}
		}

		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "22. Failed to update Appdrunit status")
		}

		//update Spec
		if bAllVmDKsActive {
			instance.Status.FailoverStatus.PowerOnStatus = RECOVERY_ACTIVITY_COMPLETED
			instance.Status.FailoverStatus.RehydrationStatus = RECOVERY_ACTIVITY_IN_PROGRESS
		}
		//fmt.Println("Sleep Over for 10 seconds after failover trigger.....")
		// Calling Sleep method
		//time.Sleep(10 * time.Second)

		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "24. Failed to update Appdrunit status")
		}

		//Requeue the work to reconsiler
		if err != nil {
			return reconcile.Result{Requeue: true}, nil
		}
	}

	if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
		reqLogger.Error(err, "25 Failed to update Appdrunit status")
	}

	if (instance.Spec.TriggerCancelRecoveryOperation) && (instance_dr != nil) {
		instance.Status.FailoverStatus.OverallFailoverStatus = RECOVERY_ACTIVITY_NOT_STARTED
		instance.Status.FailoverStatus.InfrastructureStatus = RECOVERY_ACTIVITY_NOT_STARTED
		instance.Status.FailoverStatus.RehydrationStatus = RECOVERY_ACTIVITY_NOT_STARTED
		instance.Status.FailoverStatus.PowerOnStatus = RECOVERY_ACTIVITY_NOT_STARTED
		instance.Status.FailoverStatus.PowerOffStatus = RECOVERY_ACTIVITY_NOT_STARTED
		err := CancelFailover(instance.Spec.VesToken, instance.Status.FailoverVmListStatus)
		if err != nil {
			fmt.Println("Failed to cancel failover .......", err)
		} else {
			instance.Spec.TriggerCancelRecoveryOperation = false
		}
	}

	if instance.Status.FailoverStatus.OverallFailoverStatus == RECOVERY_ACTIVITY_IN_PROGRESS {
		fmt.Println("Adding GetFailoverStatus .......")

		bIsFailoverCompleted, _ := GetFailoverStatus(instance.Spec.VesToken, instance.Status.FailoverVmListStatus)
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "26 Failed to update Appdrunit status")
		}

		vmmapList := instance.Status.FailoverVmListStatus
		for _, vmmap := range vmmapList {
			if vmmap.TriggerPowerOff {
				fmt.Println("FailOver completed for VM, so Powering OFF VM........", vmmap.VmName)
				VmPowerChange(vcenter_dr, vmmap.TargetVmUUID, false, false)
				fmt.Println("FailOver completed for VM, so Powering ON VM........", vmmap.VmName)
				VmPowerChange(vcenter_dr, vmmap.TargetVmUUID, true, false)
			} else if vmmap.TriggerReset {
				fmt.Println("Failover sentblocks nil, so Resetting VM........", vmmap.VmName)
				VmPowerChange(vcenter_dr, vmmap.TargetVmUUID, false, true)
			}
		}

		if bIsFailoverCompleted {
			instance.Status.FailoverStatus.OverallFailoverStatus = RECOVERY_ACTIVITY_COMPLETED
			instance.Status.FailoverStatus.RehydrationStatus = RECOVERY_ACTIVITY_COMPLETED
			if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
				reqLogger.Error(err, "27 Failed to update appdrunit status")
			}
		}
	}

	if instance.Status.FailbackStatus.OverallFailbackStatus == RECOVERY_ACTIVITY_IN_PROGRESS {
		//fmt.Println("Adding GetFailbackStatus .......")

		bIsFailbackCompleted, _ := GetFailbackStatus(instance.Spec.VesToken, instance.Status.FailbackVmListStatus)
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "28 Failed to update Appdrunit status")
		}

		for _, vm := range instance.Status.FailbackVmListStatus {

			//PowerOff the Source VM
			if vm.TriggerPowerOff {
				fmt.Println("FailBack: Power OFF of Source VM ", vm.VmName)
				VmPowerChange(vcenter_dr, vm.SourceVmUUID, false, false)

				for _, vmdk := range vm.VmdkStatusList {
					if vmdk.Follow_Seq != "" {
						if vmdk.RehydrationStatus != RECOVERY_ACTIVITY_COMPLETED {
							//Execute Delete of Failovers
							fmt.Println("DELETE Failover API: Follow_Seq is not null in failback for ID", vmdk.FailbackTriggerID)
							DeleteFailoverEntry(instance.Spec.VesToken, vmdk.FailbackTriggerID) //time.Sleep(10 * time.Second)
							time.Sleep(2 * time.Second)
						}
					}
				}
			}

		}

		if bIsFailbackCompleted {
			//TODO: PowerON PowerOFF VM
			vmmapList := instance.Status.FailbackVmListStatus
			for _, vmmap := range vmmapList {
				if vmmap.TriggerPowerOff {
					fmt.Println("FailBack completed for VM, so Powering OFF VM........", vmmap.VmName)
					VmPowerChange(vcenter, vmmap.TargetVmUUID, false, false)
					fmt.Println("FailBack completed for VM, so Powering ON VM........", vmmap.VmName)
					VmPowerChange(vcenter, vmmap.TargetVmUUID, true, false)
				}
			}
			//
			instance.Status.FailbackStatus.OverallFailbackStatus = RECOVERY_ACTIVITY_COMPLETED
			instance.Status.FailbackStatus.RehydrationStatus = RECOVERY_ACTIVITY_COMPLETED
			if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
				reqLogger.Error(err, "29 Failed to update Appdrunit status")
			}

			fmt.Println("Failback COMPLETED........")
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

		for _, vm := range vmDetailList {
			err := CreateVMDKsAtPostGresDB(instance.Spec.VesToken, vm)
			if err != nil {
				fmt.Println("Failed to create VMDK's at postgres .......", err)
			}
		}

		instance.Status.ProtectedVmList = vmDetailList
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "30. Failed to update Appdrunit status")
		}

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
			if err != nil {
				fmt.Println("Failed to change policy state of VM .......", err)
			}

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
				err := CreateVMDKsAtPostGresDB(instance.Spec.VesToken, vmDet)
				if err != nil {
					fmt.Println("Failed to create VMDK's at postgres .......", err)
				}
				//Trigger PowerOn for VM
				//fmt.Println("Powering ON VM :", vmDet.Name)
				//VmPowerChange(vcenter_dr, vmDet.VmUuid, true)
			}

			if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
				reqLogger.Error(err, "31. Failed to update Appdrunit status")
			}
		}

	}

	if instance.Status.ProtectedVmList != nil {

		if (instance.Status.FailbackStatus.OverallFailbackStatus != RECOVERY_ACTIVITY_IN_PROGRESS) ||
			(instance.Status.FailoverStatus.OverallFailoverStatus != RECOVERY_ACTIVITY_IN_PROGRESS) {

			//fmt.Println("Getting Initial Sync details: Calling GetVMDKsFromPostGresDB to get received blocks/IO info.......")

			VMDKListFromPostGresDResponse, _ := GetVMDKsFromPostGresDB(instance.Spec.VesToken)
			for _, vmdkmapinfo := range VMDKListFromPostGresDResponse.Data {
				for i, vminfo := range instance.Status.ProtectedVmList {
					bUpdatedVMDKInfo := false
					for j, vmdkinfo := range vminfo.Disks {
						if vmdkmapinfo.VmdkScope == vmdkinfo.AbsPath {
							//fmt.Println("Adding Received blocks .......")

							vmdkinfo.ReceivedBlocks = vmdkmapinfo.ReceivedBlocks
							vmdkinfo.ReceivedIOs = vmdkmapinfo.ReceivedIOs
							vmdkinfo.TotalBlocks = vmdkmapinfo.TotalBlocks
							bUpdatedVMDKInfo = true
							vminfo.Disks[j] = vmdkinfo
							break
						}
					}
					if bUpdatedVMDKInfo {
						instance.Status.ProtectedVmList[i] = vminfo
						if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
							reqLogger.Error(err, "32. Failed to update Appdrunit status")
						}

						break
					}
				}
			}
		}
		//fmt.Println("Sleep Over for 5 seconds.....")
		// Calling Sleep method
		//time.Sleep(5 * time.Second)
	}

	//update Spec
	if err = r.Client.Update(context.TODO(), instance); err != nil {
		reqLogger.Error(err, "35. Failed to update Appdrunit", "Site", instance.Name)
		return reconcile.Result{}, err
	}

	if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
		reqLogger.Error(err, "36. Failed to update Appdrunit status")
	}

	return reconcile.Result{RequeueAfter: time.Millisecond * 100, Requeue: true}, nil
	//return reconcile.Result{Requeue: true}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppDRUnitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&draasv1alpha1.AppDRUnit{}).
		Complete(r)
}

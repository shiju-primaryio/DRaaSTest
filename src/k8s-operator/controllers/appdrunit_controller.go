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

	if instance.Spec.ProtectVMUUIDList != nil {
		var vmDetails draasv1alpha1.VMStatus
		for _, vm := range instance.Spec.ProtectVMUUIDList {
			fmt.Println("Changing policy state for Vm: ", vm.VmUuid)

			vmDetails, err = ChangePolicyState(instance.Spec.VCenter, vm.VmUuid, vm.IsPolicyAttach)
			if err != nil {
				//reqLogger.Error(err, "Failed to change state of VM.")
				fmt.Println("Failed to attach VM .......")
			}

			bFoundVmAlreadyProtected := false
			fmt.Println("vmDetails.Name : ", vmDetails.Name)
			for _, pvm := range instance.Status.ProtectedVmList {
				if pvm.VmUuid == vmDetails.VmUuid {
					pvm = vmDetails
					bFoundVmAlreadyProtected = true
				}
			}
			if bFoundVmAlreadyProtected == false {
				instance.Status.ProtectedVmList = append(instance.Status.ProtectedVmList, vmDetails)
			}
		}
		for _, vm := range instance.Status.ProtectedVmList {
			fmt.Println("VmUUID: ", vm.VmUuid)
			fmt.Println("vmName : ", vm.Name)
		}

		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update Site status")
		}

		fmt.Println("Setting ProtectVMUUIDList to Nil .......")
		instance.Spec.ProtectVMUUIDList = nil

		if err = r.Client.Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update", "Site", instance.Name)
			return reconcile.Result{}, err
		}

	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppDRUnitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&draasv1alpha1.AppDRUnit{}).
		Complete(r)
}

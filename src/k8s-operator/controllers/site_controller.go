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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	draasv1alpha1 "github.com/CacheboxInc/DRaaS/src/k8s-operator/api/v1alpha1"
	"github.com/golang/glog"
)

var logger = log.Log.WithName("controller_site")

// SiteReconciler reconciles a Site object
type SiteReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=draas.primaryio.com,resources=sites,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=draas.primaryio.com,resources=sites/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=draas.primaryio.com,resources=sites/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Site object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *SiteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	//fmt.Println("\n ***Site: Current date and time is: ", time.Now().String())

	// TODO(user): your logic here
	reqLogger := logger.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling Site Config")

	var err error
	// Fetch the Site instance
	instance := &draasv1alpha1.Site{}
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
			glog.Errorf("Failed to update Site status : %v", err)
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

	/* if instance.Spec.VMList != nil {
		for _, vmSpec := range instance.Spec.VMList {
			// if err = powerOnOffVM(vmSpec.UUID, vmSpec.PowerOn); err != nil {
			//  glog.Errorf("Error powering ON/OFF VM : %v", err)
			// 	return reconcile.Result{}, err
			// }
			instance.Status.VmMap[vmSpec.UUID].IsProtected = vmSpec.IsPowerOn
		}
		return reconcile.Result{}, nil
	}
	*/
	//fmt.Println("\n Site: Current date and time is: ", time.Now().String())

	var policyId string
	fmt.Println("instance.Status.PolicyId: ", instance.Status.PolicyId)
	//If Host field is set, then create Storage Policy if doesn't exist already
	if instance.Spec.StoragePolicy.Host != "" && instance.Status.PolicyId == "" {
		fmt.Println("Creating Storage Policy.......")
		policyId, err = CreateStoragePolicyForSite(instance.Spec.VCenter, instance.Spec.StoragePolicy)
		if err != nil {
			//reqLogger.Error(err, "Failed to create policy.")
			fmt.Println(err, "Failed to create policy.")
		}

		instance.Status.PolicyId = policyId

		//Reset the VMList in Spec
		if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
			glog.Errorf("Failed to update Site status : %v", err)
		}
	}

	//fmt.Println("get vmList.......")
	// Fetch VMs from VCenter on Site Creation only
	vmList, err := getVmList(instance.Spec.VCenter, nil)
	if err != nil {
		//reqLogger.Error(err, "Failed to fetch VM list")
		fmt.Println("Failed to fetch VM list :")

	}

	instance.Status.VmList = vmList

	if instance.Spec.VMList != nil {
		for _, vm := range instance.Spec.VMList {
			var taskName string
			fmt.Println("Changing power state of Vm: ", vm.UUID)
			taskName, err = VmPowerChange(instance.Spec.VCenter, vm.UUID, vm.IsPowerOn, false)

			fmt.Println("VM task name :", taskName)
			if err != nil {
				fmt.Println("Failed to change power state of VM.....")

				//reqLogger.Error(err, "Failed to change VM power state.")
				instance.Status.Error.ErrorMessage = "Failed to change VM power state."
			} else {
				instance.Status.Error.ErrorMessage = ""
			}
		}
		fmt.Println("Setting VMList to Nil .......")
		instance.Spec.VMList = nil

		//Reset the VMList in Spec
		if err = r.Client.Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "Failed to update", "Site", instance.Name)
			return reconcile.Result{}, err
		}

	}

	if err = r.Client.Status().Update(context.TODO(), instance); err != nil {
		glog.Errorf("Failed to update Site status : %v", err)
		return ctrl.Result{}, err
	}

	fmt.Println("Sleep Over for 5 seconds.....")
	// Calling Sleep method
	time.Sleep(5 * time.Second)

	//return ctrl.Result{}, nil
	//return reconcile.Result{Requeue: true}, nil
	return reconcile.Result{RequeueAfter: time.Millisecond * 100, Requeue: true}, nil

	//return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SiteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&draasv1alpha1.Site{}).
		Complete(r)
}

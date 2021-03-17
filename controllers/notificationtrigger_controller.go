/*


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
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tmaxiov1alpha1 "github.com/tmax-cloud/alarm-operator/api/v1alpha1"
)

// NotificationTriggerReconciler reconciles a NotificationTrigger object
type NotificationTriggerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tmax.io.my.domain,resources=notificationtriggers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tmax.io.my.domain,resources=notificationtriggers/status,verbs=get;update;patch

func (r *NotificationTriggerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	// ctx := context.Background()
	// logger := r.Log.WithValues("Reconcile", req.NamespacedName)

	// instance := &tmaxiov1alpha1.NotificationTrigger{}
	// err := r.Client.Get(ctx, req.NamespacedName, instance)
	// if err != nil {
	// 	if errors.IsNotFound(err) {
	// 		return ctrl.Result{}, nil
	// 	}
	// 	return ctrl.Result{}, err
	// }

	return ctrl.Result{}, nil
}

func (r *NotificationTriggerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmaxiov1alpha1.NotificationTrigger{}).
		Complete(r)
}

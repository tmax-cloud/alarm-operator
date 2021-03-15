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
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tmaxiov1alpha1 "github.com/tmax-cloud/alarm-operator/api/v1alpha1"
)

// NotificationReconciler reconciles a Notification object
type NotificationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tmax.io.my.domain,resources=notifications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tmax.io.my.domain,resources=notifications/status,verbs=get;update;patch

func (r *NotificationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("notification", req.NamespacedName)

	// your logic here
	instance := &tmaxiov1alpha1.Notification{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	var action tmaxiov1alpha1.NotificationType
	if &instance.Spec.Email != nil {
		action = tmaxiov1alpha1.NotificationTypeMail
	} else if &instance.Spec.Webhook != nil {
		action = tmaxiov1alpha1.NotificationTypeWebhook
	} else if &instance.Spec.Slack != nil {
		action = tmaxiov1alpha1.NotificationTypeSlack
	} else {
		action = tmaxiov1alpha1.NotificationTypeUnknown
	}

	if err := r.updateStatus(instance, action); err != nil {
		logger.Error(err, "Failed to update status")
	}

	return ctrl.Result{}, nil
}

func (r *NotificationReconciler) updateStatus(instance *tmaxiov1alpha1.Notification, action tmaxiov1alpha1.NotificationType) error {
	original := instance.DeepCopy()
	instance.Status.Type = action

	return r.Client.Status().Patch(context.TODO(), instance, client.MergeFrom(original))
}

func (r *NotificationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmaxiov1alpha1.Notification{}).
		Complete(r)
}

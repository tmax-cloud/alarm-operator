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
	"k8s.io/apimachinery/pkg/types"
	"path"
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

// +kubebuilder:rbac:groups=alarm.tmax.io,resources=notificationtriggers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=alarm.tmax.io,resources=notificationtriggers/status,verbs=get;update;patch

func (r *NotificationTriggerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	const finalizer = "notificationtrigger.finalizer.alarm-operator.tmax.io"
	ctx := context.Background()
	logger := r.Log.WithValues("reconcile", req.NamespacedName)

	o := &tmaxiov1alpha1.NotificationTrigger{}
	err := r.Client.Get(ctx, req.NamespacedName, o)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	monitor := &tmaxiov1alpha1.Monitor{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: o.Spec.Monitor, Namespace: o.Namespace}, monitor)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if o.ObjectMeta.DeletionTimestamp.IsZero() {
		if !hasFinalizer(o.ObjectMeta, finalizer) {
			o.ObjectMeta.Finalizers = append(o.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(ctx, o); err != nil {
				return ctrl.Result{}, err
			}
		}

		path.Join(req.Namespace, req.Name)
		if hasSubscribersAnnotation(monitor.ObjectMeta, o.ObjectMeta) {
			logger.Info("already in subscribers", "subject", monitor.Name)
		} else {
			logger.Info("add subscriber", "subject", monitor.Name)
			addSubscribersAnnotation(&monitor.ObjectMeta, o.ObjectMeta)
			if err := r.Update(ctx, monitor); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if hasFinalizer(o.ObjectMeta, finalizer) {
			logger.Info("remove subscriber", "subject", monitor.Name)
			removeSubscribersAnnotation(&monitor.ObjectMeta, o.ObjectMeta)
			if err := r.Update(ctx, monitor); err != nil {
				return ctrl.Result{}, err
			}

			removeFinalizer(&o.ObjectMeta, finalizer)
			if err := r.Update(ctx, o); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

func (r *NotificationTriggerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmaxiov1alpha1.NotificationTrigger{}).
		Complete(r)
}

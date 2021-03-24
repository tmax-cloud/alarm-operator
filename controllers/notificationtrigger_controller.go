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
	"path"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tmaxiov1alpha1 "github.com/tmax-cloud/alarm-operator/api/v1alpha1"
)

const ntrFinalizer = "notificationtrigger.finalizer.alarm-operator.tmax.io"

// NotificationTriggerReconciler reconciles a NotificationTrigger object
type NotificationTriggerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tmax.io.my.domain,resources=notificationtriggers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tmax.io.my.domain,resources=notificationtriggers/status,verbs=get;update;patch

func (r *NotificationTriggerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
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

	logger.Info("process", "spec.notification", o.Spec.Notification, "spec.watchFieldPath", o.Spec.FieldPath, "monitor", o.Spec.Monitor)

	mon := &tmaxiov1alpha1.Monitor{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: o.Spec.Monitor, Namespace: o.Namespace}, mon)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if o.ObjectMeta.DeletionTimestamp.IsZero() {
		if !hasNtrFinalizer(o) {
			o.ObjectMeta.Finalizers = append(o.ObjectMeta.Finalizers, ntrFinalizer)
			if err := r.Update(context.Background(), o); err != nil {
				return ctrl.Result{}, err
			}
		}

		if hasSubscribeAnnotation(o, mon) {
			// logger.Info("Already has subscribe annotation", "Monitor", mon.Name)
		} else {
			// logger.Info("Add subscribe annotation", "Monitor", mon.Name)
			addSubscribeAnnotation(o, mon)
			if err := r.Update(context.Background(), mon); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if hasNtrFinalizer(o) {
			removeSubscribeAnnotation(o, mon)
			// logger.Info("remove subscriber annotation", "name", mon.Name, "annotation", mon.Annotations["subscribers"])
			if err := r.Update(context.Background(), mon); err != nil {
				return ctrl.Result{}, err
			}

			removeNtrFinalizer(o)
			if err := r.Update(context.Background(), o); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func hasNtrFinalizer(m *tmaxiov1alpha1.NotificationTrigger) bool {
	for _, f := range m.ObjectMeta.Finalizers {
		if f == ntrFinalizer {
			return true
		}
	}
	return false
}

func removeNtrFinalizer(m *tmaxiov1alpha1.NotificationTrigger) {
	newFinalizers := []string{}
	for _, f := range m.ObjectMeta.Finalizers {
		if f == ntrFinalizer {
			continue
		}
		newFinalizers = append(newFinalizers, f)
	}
	m.ObjectMeta.Finalizers = newFinalizers
}

func hasSubscribeAnnotation(o *tmaxiov1alpha1.NotificationTrigger, mon *tmaxiov1alpha1.Monitor) bool {
	entry := path.Join(o.Namespace, o.Name)
	subscribers := strings.Split(mon.Annotations["subscribers"], ",")
	for _, subcriber := range subscribers {
		if subcriber == entry {
			return true
		}
	}
	return false
}

func addSubscribeAnnotation(o *tmaxiov1alpha1.NotificationTrigger, mon *tmaxiov1alpha1.Monitor) {
	var newAnnotation string
	entry := path.Join(o.Namespace, o.Name)
	if mon.Annotations["subscribers"] == "" {
		newAnnotation = entry
	} else {
		newAnnotation = strings.Join([]string{mon.Annotations["subscribers"], entry}, ",")
	}
	mon.Annotations["subscribers"] = newAnnotation
}

func removeSubscribeAnnotation(o *tmaxiov1alpha1.NotificationTrigger, mon *tmaxiov1alpha1.Monitor) {
	others := []string{}
	entry := path.Join(o.Namespace, o.Name)
	subscribers := strings.Split(mon.Annotations["subscribers"], ",")
	for _, subscriber := range subscribers {
		if subscriber == entry {
			continue
		}
		others = append(others, subscriber)
	}
	mon.Annotations["subscribers"] = strings.Join(others, ",")
}

func (r *NotificationTriggerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmaxiov1alpha1.NotificationTrigger{}).
		Complete(r)
}

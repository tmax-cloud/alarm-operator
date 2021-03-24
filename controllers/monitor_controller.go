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
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tmaxiov1alpha1 "github.com/tmax-cloud/alarm-operator/api/v1alpha1"
	"github.com/tmax-cloud/alarm-operator/pkg/cron"
	"github.com/tmax-cloud/alarm-operator/pkg/monitor"
)

const monFinalizer = "monitor.finalizer.alarm-operator.tmax.io"

var s *cron.Scheduler

func init() {
	s = cron.NewScheduler(context.Background(), 100)
	go s.Start()
}

// MonitorReconciler reconciles a Monitor object
type MonitorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tmax.io.my.domain,resources=monitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tmax.io.my.domain,resources=monitors/status,verbs=get;update;patch

func (r *MonitorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("reconcile", req.NamespacedName)

	o := &tmaxiov1alpha1.Monitor{}
	err := r.Client.Get(ctx, req.NamespacedName, o)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if o.ObjectMeta.DeletionTimestamp.IsZero() {
		if !hasMonFinalizer(o) {
			o.ObjectMeta.Finalizers = append(o.ObjectMeta.Finalizers, monFinalizer)
			if err := r.Update(context.Background(), o); err != nil {
				return ctrl.Result{}, err
			}
		}

		var handler cron.TaskFunc
		task := &monitor.MonitorUpdater{
			Client: r.Client,
			Target: req.NamespacedName,
			Logger: logger,
		}
		handler = task.Handle

		subs := strings.Split(o.Annotations["subscribers"], ",")
		if len(subs) > 0 && subs[0] != "" {
			postTask := &monitor.PublishTrigger{
				Client: r.Client,
				Target: req.NamespacedName,
				Logger: logger,
			}

			handler = task.HandleFunc(postTask.Handle)
		}

		s.Schedule(o.Name).Every(o.Spec.Interval).Second().Do(handler)
	} else {
		if hasMonFinalizer(o) {
			removeMonFinalizer(o)
			s.Cancel(o.Name)
			if err := r.Update(context.Background(), o); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// FIXME:
	// o.Status.Results = append(o.Status.Results, result)

	return ctrl.Result{}, nil
}

func hasMonFinalizer(m *tmaxiov1alpha1.Monitor) bool {
	for _, f := range m.ObjectMeta.Finalizers {
		if f == monFinalizer {
			return true
		}
	}
	return false
}

func removeMonFinalizer(m *tmaxiov1alpha1.Monitor) {
	newFinalizers := []string{}
	for _, f := range m.ObjectMeta.Finalizers {
		if f == monFinalizer {
			continue
		}
		newFinalizers = append(newFinalizers, f)
	}
	m.ObjectMeta.Finalizers = newFinalizers
}

func (r *MonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmaxiov1alpha1.Monitor{}).
		Complete(r)
}

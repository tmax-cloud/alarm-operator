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
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	tmaxiov1alpha1 "github.com/tmax-cloud/alarm-operator/api/v1alpha1"
	"github.com/tmax-cloud/alarm-operator/pkg/monitor/scheduler"
)

const jobFinalizer = "monitor.finalizer.alarm-operator.tmax.io"

type ResourceFetchTask struct {
	client.Client
	target *tmaxiov1alpha1.Monitor
	url    string
	body   string
	logger logr.Logger
}

func (t *ResourceFetchTask) Run() error {

	t.logger.Info("Start Task!!")

	result := tmaxiov1alpha1.MonitorResult{}

	req, err := http.NewRequest("GET", t.url, bytes.NewBuffer([]byte(t.body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	result.LastTime = time.Now().Format(time.RFC3339)
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		result.Status = "Success"
		dat, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		result.Value = string(dat)
	} else {
		result.Status = "Fail"
	}

	t.target.Status.Result = result
	t.logger.Info("result", "status", result.Status, "value", result.Value, "time", result.LastTime)

	err = t.Status().Update(context.Background(), t.target)
	if err != nil {
		return err
	}

	t.logger.Info("END Task!!")
	return nil
}

var gRunner *scheduler.JobRunner

func init() {
	gRunner = scheduler.NewJobRunner(context.Background(), 100)
	go gRunner.Start()
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
	logger := r.Log.WithValues("monitor", req.NamespacedName)

	o := &tmaxiov1alpha1.Monitor{}
	err := r.Client.Get(ctx, req.NamespacedName, o)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	logger.Info("process", "spec.url", o.Spec.URL, "spec.body", o.Spec.Body, "interval", o.Spec.Interval)

	if o.ObjectMeta.DeletionTimestamp.IsZero() {
		if !hasFinalizer(o) {
			o.ObjectMeta.Finalizers = append(o.ObjectMeta.Finalizers, jobFinalizer)
			if err := r.Update(context.Background(), o); err != nil {
				return ctrl.Result{}, err
			}
		}

		task := &ResourceFetchTask{
			url:    o.Spec.URL,
			body:   o.Spec.Body,
			Client: r.Client,
			target: o,
			logger: logger,
		}

		job := scheduler.NewIntervalJob(o.Name, time.Duration(o.Spec.Interval), task)
		gRunner.Schedule(job)
	} else {
		if hasFinalizer(o) {
			removeFinalizer(o)
			gRunner.CancelJob(o.Name)
			if err := r.Update(context.Background(), o); err != nil {
				return reconcile.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// FIXME:
	// o.Status.Results = append(o.Status.Results, result)

	return ctrl.Result{}, nil
}

func hasFinalizer(m *tmaxiov1alpha1.Monitor) bool {
	for _, f := range m.ObjectMeta.Finalizers {
		if f == jobFinalizer {
			return true
		}
	}
	return false
}

func removeFinalizer(m *tmaxiov1alpha1.Monitor) {
	newFinalizers := []string{}
	for _, f := range m.ObjectMeta.Finalizers {
		if f == jobFinalizer {
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

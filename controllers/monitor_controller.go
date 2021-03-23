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

	tmaxiov1alpha1 "github.com/tmax-cloud/alarm-operator/api/v1alpha1"
)

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

	fetchReq, err := http.NewRequest("GET", o.Spec.URL, bytes.NewBuffer([]byte(o.Spec.Body)))
	if err != nil {
		logger.Error(err, "failed to make request", "url", o.Spec.URL)
		return ctrl.Result{RequeueAfter: requeueDuration}, err
	}
	fetchReq.Header.Set("Content-Type", "application/json")

	fetchRes, err := http.DefaultClient.Do(fetchReq)
	if err != nil {
		logger.Error(err, "failed to fetch data", "url", o.Spec.URL)
		return ctrl.Result{RequeueAfter: requeueDuration}, err
	}

	result := tmaxiov1alpha1.MonitorResult{
		LastTime: time.Now().Format(time.RFC3339),
	}

	logger.Info("resposne", "statuscode", fetchRes.StatusCode, "status", fetchRes.Status)
	if fetchRes.StatusCode >= 200 && fetchRes.StatusCode < 300 {
		result.Status = "Success"
		dat, err := ioutil.ReadAll(fetchRes.Body)
		if err != nil {
			logger.Error(err, "failed to read data", "url", o.Spec.URL)
			return ctrl.Result{RequeueAfter: requeueDuration}, err
		}
		defer fetchRes.Body.Close()
		result.Value = string(dat)
	} else {
		result.Status = "Fail"
	}

	result.Status = "Success"
	logger.Info("result", "status", result.Status, "value", result.Value, "time", result.LastTime)

	// FIXME:
	// o.Status.Results = append(o.Status.Results, result)
	o.Status.Result = result

	err = r.Client.Status().Update(ctx, o)
	if err != nil {
		logger.Error(err, "failed to update monitor")
		return ctrl.Result{RequeueAfter: requeueDuration}, err
	}

	logger.Info("end")
	return ctrl.Result{}, nil
}

func (r *MonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmaxiov1alpha1.Monitor{}).
		Complete(r)
}

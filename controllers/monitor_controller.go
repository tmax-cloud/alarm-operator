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

import "C"
import (
	"bytes"
	"context"
	"fmt"
	"github.com/Jeffail/gabs"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tmaxiov1alpha1 "github.com/tmax-cloud/alarm-operator/api/v1alpha1"
	"github.com/tmax-cloud/alarm-operator/pkg/cron"
)

var s *cron.Scheduler
var httpcli *http.Client

func init() {
	httpcli = &http.Client{Transport: http.DefaultTransport}
	s = cron.NewScheduler(context.Background(), 100)
	go s.Start()
}

// MonitorReconciler reconciles a Monitor object
type MonitorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=alarm.tmax.io,resources=monitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=alarm.tmax.io,resources=monitors/status,verbs=get;update;patch

func (r *MonitorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	const finalizer = "monitor.finalizer.alarm-operator.tmax.io"
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

	if !o.ObjectMeta.DeletionTimestamp.IsZero() {
		if hasFinalizer(o.ObjectMeta, finalizer) {
			removeFinalizer(&o.ObjectMeta, finalizer)
			s.Schedule(o.Name).Cancel()
			if err := r.Update(ctx, o); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if !hasFinalizer(o.ObjectMeta, finalizer) {
		addFinalizer(&o.ObjectMeta, finalizer)
		if err := r.Update(ctx, o); err != nil {
			return ctrl.Result{}, err
		}
	}

	s.Schedule(o.Name).Every(o.Spec.Interval).Second().Do(func(ctx context.Context) error {
		retCode, dat, err := fetchResource(ctx, o.Spec.URL, []byte(o.Spec.Body))
		result := tmaxiov1alpha1.MonitorResult{
			Status:    "Success",
			Value:     string(dat),
			UpdatedAt: time.Now().Format(time.RFC3339),
		}
		if err != nil || retCode >= 300 {
			result.Status = "Fail"
		}

		latestIdx := len(o.Status.History) - 1
		if len(o.Status.History) > 0 && len(o.Status.History[latestIdx].Value) > tmaxiov1alpha1.ValueSizeLimit {
			o.Status.History[latestIdx].Value = tmaxiov1alpha1.ValueReplacement
		}
		o.Status.History = append(o.Status.History, result)
		if len(o.Status.History) > tmaxiov1alpha1.HistoryLimit {
			start := len(o.Status.History) - tmaxiov1alpha1.HistoryLimit
			o.Status.History = o.Status.History[start:]
		}

		logger.Info("Update", "value", result.Status)
		if err = r.Status().Update(ctx, o); err != nil {
			return err
		}

		// FIXME: next trigger handling logic must be seperated.
		subscribers := parseSubscribers(o)
		for _, s := range subscribers {
			nt := &tmaxiov1alpha1.NotificationTrigger{}
			if err := r.Get(ctx, s, nt); err != nil {
				logger.Error(err, "failed to get notification trigger")
				return err
			}

			jsonParsed, err := gabs.ParseJSON(dat)
			if err != nil {
				return err
			}
			v := jsonParsed.Path(nt.Spec.FieldPath).Data()
			logger.Info("parsed field", "value", v)

			result := tmaxiov1alpha1.NotificationTriggerResult{}
			if eval(v, nt.Spec.Operand, nt.Spec.Op) {
				n := &tmaxiov1alpha1.Notification{}
				if err := r.Get(ctx, types.NamespacedName{Namespace: nt.Namespace, Name: nt.Spec.Notification}, n); err != nil {
					result.Message = "failed to get notification from resource"
					logger.Error(err, result.Message)
				}
				if err = sendNotification(*n); err != nil {
					result.Message = "failed to send notification"
					logger.Error(err, result.Message)
				}
				result.Triggered = true
				result.UpdatedAt = time.Now().Format(time.RFC3339)
			} else {
				result.Triggered = false
				result.Message = fmt.Sprintf("condition not matched")
			}

			nt.Status.History = append(nt.Status.History, result)
			if len(nt.Status.History) > tmaxiov1alpha1.HistoryLimit {
				start := len(nt.Status.History) - tmaxiov1alpha1.HistoryLimit
				nt.Status.History = nt.Status.History[start:]
			}

			if err := r.Status().Update(ctx, nt); err != nil {
				return err
			}
		}

		return nil
	})

	return ctrl.Result{}, nil
}

func fetchResource(ctx context.Context, url string, body []byte) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, bytes.NewBuffer(body))
	if err != nil {
		return http.StatusBadRequest, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := httpcli.Do(req)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	defer response.Body.Close()
	body, err = ioutil.ReadAll(response.Body)
	return response.StatusCode, body, err
}

func parseSubscribers(o *tmaxiov1alpha1.Monitor) []types.NamespacedName {
	ret := []types.NamespacedName{}
	subscribers := strings.Split(o.Annotations["subscribers"], ",")
	for _, subscriber := range subscribers {
		if subscriber == "" {
			continue
		}
		tokens := strings.Split(subscriber, "/")
		ret = append(ret, types.NamespacedName{Namespace: tokens[0], Name: tokens[1]})
	}
	return ret
}

func eval(op1 interface{}, op2 string, op string) bool {
	switch op {
	case "gt", "<":
		switch op1 := op1.(type) {
		case int:
			operand, _ := strconv.Atoi(op2)
			if op1 > operand {
				return true
			}
		case float64:
			operand, _ := strconv.Atoi(op2)
			if op1 > float64(operand) {
				return true
			}
		case string:
			if op1 > op2 {
				return true
			}
		}
	case "gte", "<=":
		switch op1 := op1.(type) {
		case int:
			operand, _ := strconv.Atoi(op2)
			if op1 >= operand {
				return true
			}
		case float64:
			operand, _ := strconv.Atoi(op2)
			if op1 >= float64(operand) {
				return true
			}
		case string:
			if op1 >= op2 {
				return true
			}
		}
	case "eq", "==":
		switch op1 := op1.(type) {
		case int:
			operand, _ := strconv.Atoi(op2)
			if op1 == operand {
				return true
			}
		case float64:
			operand, _ := strconv.Atoi(op2)
			if op1 == float64(operand) {
				return true
			}
		case string:
			if op1 == op2 {
				return true
			}
		}
	case "lte", ">=":
		switch op1 := op1.(type) {
		case int:
			operand, _ := strconv.Atoi(op2)
			if op1 <= operand {
				return true
			}
		case float64:
			operand, _ := strconv.Atoi(op2)
			if op1 <= float64(operand) {
				return true
			}
		case string:
			if op1 <= op2 {
				return true
			}
		}
	case "lt", ">":
		switch op1 := op1.(type) {
		case int:
			operand, _ := strconv.Atoi(op2)
			if op1 < operand {
				return true
			}
		case float64:
			operand, _ := strconv.Atoi(op2)
			if op1 < float64(operand) {
				return true
			}
		case string:
			if op1 < op2 {
				return true
			}
		}
	}

	return false
}

func sendNotification(o tmaxiov1alpha1.Notification) error {
	if o.Status.EndPoint == "" {
		return fmt.Errorf("notification's endpoint not prepared")
	}

	req, err := http.NewRequest("POST", o.Status.EndPoint, bytes.NewBuffer([]byte("")))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", o.Status.ApiKey)

	if _, err = http.DefaultClient.Do(req); err != nil {
		return err
	}
	return nil
}

func (r *MonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmaxiov1alpha1.Monitor{}).
		Complete(r)
}

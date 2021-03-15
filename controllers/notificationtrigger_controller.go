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
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tmaxiov1alpha1 "github.com/tmax-cloud/alarm-operator/api/v1alpha1"
	"github.com/tmax-cloud/alarm-operator/pkg/notification"
	"github.com/tmax-cloud/alarm-operator/pkg/notification/datasource"
)

var registry *notification.NotificationRegistry

func init() {

	redisURL, _ := url.Parse(os.Getenv("REDIS_URL"))
	redisPort, _ := strconv.Atoi(strings.TrimPrefix(redisURL.Path, "/"))

	ds, err := datasource.NewRedisDataSource(redisURL.Host, "", redisPort)
	if err != nil {
		panic(err)
	}
	registry = notification.NewNotificationRegistry(ds)
}

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
	logger := r.Log.WithValues("notification", req.NamespacedName)

	// your logic here
	instance := &tmaxiov1alpha1.NotificationTrigger{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	logger.Info("NotificationTrigger", "name", instance.Name, "Notification", instance.Spec.Notification)

	action := &tmaxiov1alpha1.Notification{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: instance.Spec.Notification, Namespace: req.Namespace}, action)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{}, err
	}

	var noti notification.Notification
	if &action.Spec.Email != nil {
		noti = notification.MailNotification{
			SMTPConnection: notification.SMTPConnection{
				Host: "smtp.gmail.com",
				Port: 587,
			},
			SMTPAccount: notification.SMTPAccount{
				Username: "voidmain0805@gmail.com",
				Password: "jmmdhkfzeuyivtmb",
			},
			MailMessage: notification.MailMessage{
				From:    action.Spec.Email.From,
				To:      action.Spec.Email.To,
				Subject: action.Spec.Email.Subject,
				Body:    action.Spec.Email.Body,
			},
		}
	} else if &action.Spec.Webhook != nil {
		// TODO:
	} else if &action.Spec.Slack != nil {
		// TODO:
	} else {
		// TODO:
	}

	if err := registry.Register(instance.Name, noti); err != nil {
		logger.Error(err, "Failed to register notification")
		return ctrl.Result{}, err
	}

	if err := r.updateStatus(instance); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NotificationTriggerReconciler) updateStatus(instance *tmaxiov1alpha1.NotificationTrigger) error {
	original := instance.DeepCopy()

	id := instance.Spec.Notification
	host := "notification-server:8080"
	instance.Status.EndPoint = fmt.Sprintf("http://%s.%s", id, host)

	return r.Client.Status().Patch(context.TODO(), instance, client.MergeFrom(original))
}

func (r *NotificationTriggerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmaxiov1alpha1.NotificationTrigger{}).
		Complete(r)
}

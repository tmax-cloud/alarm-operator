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
	corev1 "k8s.io/api/core/v1"
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
	logger := r.Log.WithValues("Reconcile", req.NamespacedName)

	o := &tmaxiov1alpha1.Notification{}
	err := r.Client.Get(ctx, req.NamespacedName, o)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	var noti notification.Notification
	if &o.Spec.Email != nil {

		smtpcfg := &tmaxiov1alpha1.SMTPConfig{}
		// XXX: Is SMTPConfig should be in same namespace?
		err = r.Client.Get(ctx, types.NamespacedName{Name: o.Spec.Email.SMTPConfig, Namespace: req.Namespace}, smtpcfg)
		if err != nil {
			if errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}

		smtpSecret := &corev1.Secret{}
		// XXX: Is Secret should be in same namespace?
		err = r.Client.Get(ctx, types.NamespacedName{Name: smtpcfg.Spec.CredentialSecret, Namespace: req.Namespace}, smtpSecret)
		if err != nil {
			if errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}

		var username string
		var password string

		switch smtpSecret.Type {
		case corev1.SecretTypeOpaque, corev1.SecretTypeBasicAuth:
			bUsername, ok := smtpSecret.Data[corev1.BasicAuthUsernameKey]
			if !ok {
				return ctrl.Result{}, fmt.Errorf("username not found: %s\n", corev1.BasicAuthUsernameKey)
			}
			username = string(bUsername)
			bPassword, ok := smtpSecret.Data[corev1.BasicAuthPasswordKey]
			if !ok {
				return ctrl.Result{}, fmt.Errorf("password not found: %s\n", corev1.BasicAuthPasswordKey)
			}
			password = string(bPassword)
		default:
			// TODO: load from controller configmap
		}

		noti = notification.MailNotification{
			SMTPConnection: notification.SMTPConnection{
				Host: smtpcfg.Spec.Host,
				Port: smtpcfg.Spec.Port,
			},
			SMTPAccount: notification.SMTPAccount{
				Username: username,
				Password: password,
			},
			MailMessage: notification.MailMessage{
				From:    o.Spec.Email.From,
				To:      o.Spec.Email.To,
				Subject: o.Spec.Email.Subject,
				Body:    o.Spec.Email.Body,
			},
		}
	} else if &o.Spec.Webhook != nil {
		// TODO:
	} else if &o.Spec.Slack != nil {
		// TODO:
	} else {
		// TODO:
	}

	if err := registry.Register(o.Name, noti); err != nil {
		logger.Error(err, "Failed to register notification")
		return ctrl.Result{}, err
	}

	if err := r.updateStatus(o); err != nil {
		logger.Error(err, "Failed to update trigger")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NotificationReconciler) GetEndpoint(id string) (string, error) {
	ctx := context.Background()

	o := &corev1.Service{}
	// FIXME: Designate namespacename from variable
	err := r.Client.Get(ctx, types.NamespacedName{Name: "notifier", Namespace: "alarm-operator-system"}, o)
	if err != nil {
		return "", err
	}

	// FIXME:
	var port int32
	for _, p := range o.Spec.Ports {
		if p.Name == "http" {
			port = p.Port
		}
	}

	return fmt.Sprintf("http://%s.%s.xip.io:%d", id, o.Spec.ClusterIP, port), nil
}

func (r *NotificationReconciler) updateStatus(o *tmaxiov1alpha1.Notification) error {

	if &o.Spec.Email != nil {
		o.Status.Type = tmaxiov1alpha1.NotificationTypeMail
	} else if &o.Spec.Webhook != nil {
		o.Status.Type = tmaxiov1alpha1.NotificationTypeWebhook
	} else if &o.Spec.Slack != nil {
		o.Status.Type = tmaxiov1alpha1.NotificationTypeSlack
	} else {
		o.Status.Type = tmaxiov1alpha1.NotificationTypeUnknown
	}

	endpoint, err := r.GetEndpoint(o.Name)
	if err != nil {
		return err
	}
	o.Status.EndPoint = endpoint

	return r.Status().Update(context.TODO(), o)
}

func (r *NotificationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmaxiov1alpha1.Notification{}).
		Complete(r)
}

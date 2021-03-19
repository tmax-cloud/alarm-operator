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
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tmaxiov1alpha1 "github.com/tmax-cloud/alarm-operator/api/v1alpha1"
	"github.com/tmax-cloud/alarm-operator/pkg/notification"
	notifiercli "github.com/tmax-cloud/alarm-operator/pkg/notifier/client"
)

const requeueDuration = time.Second * 3

var notifier *notifiercli.Notifier

func init() {

	notifier = notifiercli.New(os.Getenv("NOTIFIER_URL"),
		&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		})

}

// NotificationReconciler reconciles a Notification object
type NotificationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tmax.io.my.domain,resources=notifications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tmax.io.my.domain,resources=notifications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tmax.io.my.domain,resources=smtpconfigs,verbs=get;list;watch;
// +kubebuilder:rbac:groups=tmax.io.my.domain,resources=smtpconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;

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

	notiType, noti, err := r.getNotificationFromResource(ctx, o)
	if err != nil {
		logger.Error(err, "Failed to generate notification object from resource")
		return ctrl.Result{RequeueAfter: requeueDuration}, err
	}

	resp, err := notifier.Register(o.Name, notiType, noti)
	if err != nil {
		logger.Error(err, "Failed to register notification")
		return ctrl.Result{RequeueAfter: requeueDuration}, err
	}

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err, "Failed to read response from notifier")
		return ctrl.Result{}, err
	}

	defer resp.Body.Close()
	logger.Info("Registered ", "response", string(ret))

	if err := r.updateStatus(ctx, o); err != nil {
		logger.Error(err, "Failed to update trigger")
		return ctrl.Result{RequeueAfter: requeueDuration}, err
	}

	return ctrl.Result{}, nil
}

func (r *NotificationReconciler) getNotificationFromResource(ctx context.Context, o *tmaxiov1alpha1.Notification) (string, notification.Notification, error) {

	var ret notification.Notification
	var rtype string

	if &o.Spec.Email != nil {

		smtpcfg := &tmaxiov1alpha1.SMTPConfig{}
		// XXX: Is SMTPConfig should be in same namespace?
		err := r.Client.Get(ctx, types.NamespacedName{Name: o.Spec.Email.SMTPConfig, Namespace: o.Namespace}, smtpcfg)
		if err != nil {
			if errors.IsNotFound(err) {
				return "", nil, err
			}
			return "", nil, err
		}

		smtpSecret := &corev1.Secret{}
		// XXX: Is Secret should be in same namespace?
		err = r.Client.Get(ctx, types.NamespacedName{Name: smtpcfg.Spec.CredentialSecret, Namespace: o.Namespace}, smtpSecret)
		if err != nil {
			if errors.IsNotFound(err) {
				return "", nil, err
			}
			return "", nil, err
		}

		var username string
		var password string

		switch smtpSecret.Type {
		case corev1.SecretTypeOpaque, corev1.SecretTypeBasicAuth:
			bUsername, ok := smtpSecret.Data[corev1.BasicAuthUsernameKey]
			if !ok {
				return "", nil, fmt.Errorf("username not found: %s", corev1.BasicAuthUsernameKey)
			}
			username = string(bUsername)
			bPassword, ok := smtpSecret.Data[corev1.BasicAuthPasswordKey]
			if !ok {
				return "", nil, fmt.Errorf("password not found: %s", corev1.BasicAuthPasswordKey)
			}
			password = string(bPassword)
		default:
			// TODO: load from controller configmap
		}

		rtype = "email"
		ret = notification.MailNotification{
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

	return rtype, ret, nil
}

func (r *NotificationReconciler) updateStatus(ctx context.Context, o *tmaxiov1alpha1.Notification) error {

	if &o.Spec.Email != nil {
		o.Status.Type = tmaxiov1alpha1.NotificationTypeMail
	} else if &o.Spec.Webhook != nil {
		o.Status.Type = tmaxiov1alpha1.NotificationTypeWebhook
	} else if &o.Spec.Slack != nil {
		o.Status.Type = tmaxiov1alpha1.NotificationTypeSlack
	} else {
		o.Status.Type = tmaxiov1alpha1.NotificationTypeUnknown
	}

	u, err := url.Parse(os.Getenv("NOTIFIER_URL"))
	if err != nil {
		return err
	}

	epHost := u.Hostname()
	if u.Hostname() == "localhost" {
		epHost = "127.0.0.1"
	}

	o.Status.EndPoint = fmt.Sprintf("http://%s.%s.xip.io:%s", o.Name, epHost, u.Port())
	r.Log.Info("Update", "Endpoint", o.Status.EndPoint, "Type", o.Status.Type)

	return r.Status().Update(ctx, o)
}

func (r *NotificationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmaxiov1alpha1.Notification{}).
		Complete(r)
}

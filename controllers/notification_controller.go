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
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
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

	notifier = notifiercli.New(os.Getenv("NOTIFIER_URL"))
}

// NotificationReconciler reconciles a Notification object
type NotificationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=alarm.tmax.io,resources=notifications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=alarm.tmax.io,resources=notifications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=alarm.tmax.io,resources=smtpconfigs,verbs=get;list;watch;
// +kubebuilder:rbac:groups=alarm.tmax.io,resources=smtpconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;

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

	id := extractId(o.Name, o.Namespace)

	resp, err := notifier.Register(id, notiType, noti)
	if err != nil {
		logger.Error(err, "Failed to register notification")
		return ctrl.Result{RequeueAfter: requeueDuration}, err
	}

	dat, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err, "Failed to read response from notifier")
		return ctrl.Result{}, err
	}

	defer resp.Body.Close()
	logger.Info("Registered ", "name", o.Name, "type", notiType, "apikey", dat)

	o.Status.ApiKey = string(dat)
	if err := r.updateStatus(ctx, o); err != nil {
		logger.Error(err, "Failed to update trigger")
		return ctrl.Result{RequeueAfter: requeueDuration}, err
	}

	return ctrl.Result{}, nil
}

func (r *NotificationReconciler) getNotificationFromResource(ctx context.Context, o *tmaxiov1alpha1.Notification) (string, notification.Notification, error) {
	var ret notification.Notification
	var rtype string

	if o.Spec.Email.SMTPConfig != "" {
		smtpcfg := &corev1.ConfigMap{}
		// XXX: Is SMTPConfig should be in same namespace?
		err := r.Client.Get(ctx, types.NamespacedName{Name: o.Spec.Email.SMTPConfig, Namespace: o.Namespace}, smtpcfg)
		if err != nil {
			if errors.IsNotFound(err) {
				return "", nil, err
			}
			return "", nil, err
		}

		rtype = "email"
		port, _ := strconv.Atoi(smtpcfg.Data["port"])
		ret = notification.MailNotification{
			SMTPConnection: notification.SMTPConnection{
				Host: smtpcfg.Data["host"],
				Port: port,
			},
			SMTPAccount: notification.SMTPAccount{
				Username: smtpcfg.Data["username"],
				Password: smtpcfg.Data["password"],
			},
			MailMessage: notification.MailMessage{
				From:    o.Spec.Email.From,
				To:      o.Spec.Email.To,
				Subject: o.Spec.Email.Subject,
				Body:    o.Spec.Email.Body,
			},
		}

	} else if o.Spec.Webhook.Url != "" {
		// TODO:
	} else if o.Spec.Slack.Channel != "" {
		rtype = "slack"
		ret = notification.SlackNotification{
			Authorization:  o.Spec.Slack.Authorization,
			SlackMessage: notification.SlackMessage{
				Channel:  o.Spec.Slack.Channel,
				Text:     o.Spec.Slack.Text,
			},
		}
	} else {
		// TODO:
	}

	return rtype, ret, nil
}

func (r *NotificationReconciler) updateStatus(ctx context.Context, o *tmaxiov1alpha1.Notification) error {
	if o.Spec.Email.SMTPConfig != "" {
		o.Status.Type = tmaxiov1alpha1.NotificationTypeMail
	} else if o.Spec.Webhook.Url != "" {
		o.Status.Type = tmaxiov1alpha1.NotificationTypeWebhook
	} else if o.Spec.Slack.Channel != "" {
		o.Status.Type = tmaxiov1alpha1.NotificationTypeSlack
	} else {
		o.Status.Type = tmaxiov1alpha1.NotificationTypeUnknown
	}
	loadIP := os.Getenv("LOADBALANCER_IP")
	// u, err := url.Parse(os.Getenv("NOTIFIER_URL"))
	// if err != nil {
	// 	return err
	// }
	// epHost := u.Hostname()
	// if !IsIpv4Regex(u.Hostname()) {
	// 	ips, _ := net.LookupIP(u.Hostname())
	// 	for _, ip := range ips {
	// 		epHost = ip.String()
	// 	}
	// }
	o.Status.EndPoint = fmt.Sprintf("http://alarm-ingress.%s.nip.io", loadIP)
	o.Status.Id = fmt.Sprintf("%s-%s", o.Name, o.Namespace)
	r.Log.Info("Update", "Endpoint", o.Status.EndPoint, "Type", o.Status.Type, "Id", o.Status.Id)

	return r.Status().Update(ctx, o)
}

func (r *NotificationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmaxiov1alpha1.Notification{}).
		Complete(r)
}

var ipRegex, _ = regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)

func IsIpv4Regex(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")
	return ipRegex.MatchString(ipAddress)
}

func extractId(name string, namespace string) string {
	id := name + "-" + namespace
	return id
}

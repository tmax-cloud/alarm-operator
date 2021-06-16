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
	"os"

	"github.com/go-logr/logr"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/networking/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// IngressReconciler reconciles a Ingress object
type IngressReconciler struct {
	Client v1beta1.IngressInterface
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	ingName = "alarm-operator-ingress"
)

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ingress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.6.4/pkg/reconcile
func (r *IngressReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("ingress", req.NamespacedName)
	
	ingCli, err := newIngressClient()
	r.Client = ingCli
	if err != nil {
		logger.Error(err, "Fail to generate ingress client")
		return ctrl.Result{RequeueAfter: requeueDuration}, err
	}

	o, err := r.Client.Get(ctx, ingName, metav1.GetOptions{})
	if err != nil {
		logger.Error(err, "Fail to retrieve ingress")
		return ctrl.Result{RequeueAfter: requeueDuration}, err
	}

	if len(o.Status.LoadBalancer.Ingress) > 0 && o.Status.LoadBalancer.Ingress[0].IP != "" {
		ip := o.Status.LoadBalancer.Ingress[0].IP
		os.Setenv("LOADBALANCER_IP", ip)
		if len(o.Spec.Rules) == 0 {
			return ctrl.Result{RequeueAfter: requeueDuration}, fmt.Errorf("rules for ingress are not set")
		}
		if o.Spec.Rules[0].Host != "waiting.for.loadbalancer" {
			logger.Info("Current external hostname is " + o.Spec.Rules[0].Host)
			return ctrl.Result{}, nil
		} else {
			hostname := fmt.Sprintf("alarm-ingress.%s.nip.io", ip)
			o.Spec.Rules[0].Host = hostname
			logger.Info("Current external hostname is " + o.Spec.Rules[0].Host)
		}
		if _, err := r.Client.Update(ctx, o, metav1.UpdateOptions{}); err != nil {
			logger.Error(err, "Failed to update ingress client")
		}
		return ctrl.Result{}, nil
	}
	
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1beta1.Ingress{}).
		Complete(r)
}

func newIngressClient() (v1beta1.IngressInterface, error) {
	conf, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	namespace := "alarm-operator-system"

	return clientSet.NetworkingV1beta1().Ingresses(namespace), nil
}
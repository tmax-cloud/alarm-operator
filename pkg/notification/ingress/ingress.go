package ingress

import (
	"context"
	"fmt"
	"time"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	ingName = "alarm-operator-ingress"
)

type AlarmIngressClient struct {
	client v1.IngressInterface
	id     string
}

func NewAlarmIngressClient(id string) (*AlarmIngressClient, error) {
	c, err := newIngressClient("alarm-operator-system")
	if err != nil {
		return nil, err
	}
	return &AlarmIngressClient{
		client: c,
		id:     id,
	}, nil
}

func (i *AlarmIngressClient) AddIngress(c chan error) error {
	for {
		ing, err := i.getAlarmIngress()
		if err != nil {
			c <- err
			return err
		}
		ip, err := CheckLoadBalancer(ing)
		if err == nil {
			addNewRules(ing, ip, i.id)
			if _, err := i.client.Update(context.Background(), ing, metav1.UpdateOptions{}); err != nil {
				c <- err
				return err
			}
			break
		} else {
			time.Sleep(time.Second)
		}

	}
	c <- nil
	return nil
}

func newIngressClient(namespace string) (v1.IngressInterface, error) {
	conf, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	return clientSet.NetworkingV1().Ingresses(namespace), nil
}

func (i *AlarmIngressClient) getAlarmIngress() (*networkingv1.Ingress, error) {
	return i.client.Get(context.Background(), ingName, metav1.GetOptions{})
}

func CheckLoadBalancer(ing *networkingv1.Ingress) (string, error) {
	if len(ing.Status.LoadBalancer.Ingress) > 0 && ing.Status.LoadBalancer.Ingress[0].IP != "" {
		ip := ing.Status.LoadBalancer.Ingress[0].IP
		return ip, nil
	}
	return "", fmt.Errorf("alarm ingress does not have loadbalancer")
}

func addNewRules(ing *networkingv1.Ingress, ip string, id string) {
	prefixPathType := networkingv1.PathTypePrefix
	targetHost := id + "." + ip + ".nip.io"
	defaultHost := "waiting.for.loadbalancer"
	newRules := []networkingv1.IngressRule{}
	for _, rule := range ing.Spec.Rules {
		if rule.Host != targetHost && rule.Host != defaultHost {
			newRules = append(newRules, rule)
		}
	}
	newRules = append(newRules, networkingv1.IngressRule{
		Host: targetHost,
		IngressRuleValue: networkingv1.IngressRuleValue{
			HTTP: &networkingv1.HTTPIngressRuleValue{
				Paths: []networkingv1.HTTPIngressPath{
					{
						Path:     "/",
						PathType: &prefixPathType,
						Backend: networkingv1.IngressBackend{
							Service: &networkingv1.IngressServiceBackend{
								Name: "alarm-operator-notifier",
								Port: networkingv1.ServiceBackendPort{
									Number: 8080,
								},
							},
						},
					},
				},
			},
		},
	})
	ing.Spec.Rules = newRules
}

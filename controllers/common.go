package controllers

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path"
	"strings"
)

const subscribersAnnotationKey = "subscribers"

func hasFinalizer(o v1.ObjectMeta, name string) bool {
	for _, f := range o.Finalizers {
		if f == name {
			return true
		}
	}
	return false
}

func addFinalizer(o *v1.ObjectMeta, name string) {
	o.Finalizers = append(o.Finalizers, name)
}

func removeFinalizer(o *v1.ObjectMeta, name string) {
	replacement := []string{}
	for _, f := range o.Finalizers {
		if f == name {
			continue
		}
		replacement = append(replacement, f)
	}
	o.Finalizers = replacement
}

func getSubscribersAnnotation(obj v1.ObjectMeta) []string {
	return strings.Split(obj.Annotations[subscribersAnnotationKey], ",")
}

func hasSubscribersAnnotation(obj v1.ObjectMeta, subscriber v1.ObjectMeta) bool {
	target := path.Join(subscriber.Namespace, subscriber.Name)
	for _, e := range getSubscribersAnnotation(obj) {
		if target == e {
			return true
		}
	}
	return false
}

func addSubscribersAnnotation(obj *v1.ObjectMeta, subscriber v1.ObjectMeta) {
	existings := getSubscribersAnnotation(*obj)
	existings = append(existings, path.Join(subscriber.Namespace, subscriber.Name))
	obj.Annotations[subscribersAnnotationKey] = strings.Join(existings, ",")
}

func removeSubscribersAnnotation(obj *v1.ObjectMeta, subscriber v1.ObjectMeta) {
	replacement := []string{}
	target := path.Join(subscriber.Namespace, subscriber.Name)
	for _, e := range getSubscribersAnnotation(*obj) {
		if e == target {
			continue
		}
		replacement = append(replacement, e)
	}
	obj.Annotations[subscribersAnnotationKey] = strings.Join(replacement, ",")
}

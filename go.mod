module github.com/tmax-cloud/alarm-operator

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/go-redis/redis/v7 v7.4.0
	github.com/gorilla/mux v1.8.0
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	go.uber.org/zap v1.10.0
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
)

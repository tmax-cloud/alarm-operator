package monitor

import (
	"context"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tmaxiov1alpha1 "github.com/tmax-cloud/alarm-operator/api/v1alpha1"
	"github.com/tmax-cloud/alarm-operator/pkg/cron"
)

type PublishTrigger struct {
	client.Client
	target types.NamespacedName
	logger logr.Logger
}

func (t *PublishTrigger) Handle(ctx context.Context) error {

	logger := t.logger.WithName("PublishTrigger")

	o := &tmaxiov1alpha1.Monitor{}
	err := t.Client.Get(ctx, t.target, o)
	if err != nil {
		return err
	}

	subscribers := parseSubscribers(o)
	for _, subscriber := range subscribers {
		ntr := &tmaxiov1alpha1.NotificationTrigger{}
		err := t.Client.Get(context.Background(), subscriber, ntr)
		if err != nil {
			return err
		}

		logger.Info("process", "spec.notification", ntr.Spec.NotificationName, "spec.watchFieldPath", ntr.Spec.WatchFieldPath, "monitor", ntr.Spec.MonitorName)

		jsonParsed, err := gabs.ParseJSON(cron.DataFrom(ctx))
		if err != nil {
			panic(err)
		}

		v := jsonParsed.Path(ntr.Spec.WatchFieldPath).Data()

		switch v.(type) {
		case int, int32, int64:
			logger.Info("parsed value", "field", ntr.Spec.WatchFieldPath, "value", v)
		case float32, float64:
			logger.Info("parsed value", "field", ntr.Spec.WatchFieldPath, "value", v)
		case string:
			logger.Info("parsed value", "field", ntr.Spec.WatchFieldPath, "value", v)
		case map[string]interface{}:
			logger.Info("parsed value", "field", ntr.Spec.WatchFieldPath, "value", v)
		}
	}

	return nil
}

func parseSubscribers(o *tmaxiov1alpha1.Monitor) []types.NamespacedName {
	ret := []types.NamespacedName{}
	subscribers := strings.Split(o.Annotations["subscribers"], ",")
	for _, subscriber := range subscribers {
		tokens := strings.Split(subscriber, "/")
		ret = append(ret, types.NamespacedName{Namespace: tokens[0], Name: tokens[1]})
	}
	return ret
}

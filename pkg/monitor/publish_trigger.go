package monitor

import (
	"context"
	"fmt"
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
	Target types.NamespacedName
	Logger logr.Logger
}

func (t *PublishTrigger) Handle(ctx context.Context) error {

	logger := t.Logger.WithName("PublishTrigger")

	o := &tmaxiov1alpha1.Monitor{}
	err := t.Client.Get(ctx, t.Target, o)
	if err != nil {
		return err
	}

	subscribers := parseSubscribers(o)
	if len(subscribers) == 0 {
		err := fmt.Errorf("no subscriber recorded")
		logger.Error(err, "err")
		return err
	}

	for _, subscriber := range subscribers {
		ntr := &tmaxiov1alpha1.NotificationTrigger{}
		err := t.Client.Get(context.Background(), subscriber, ntr)
		if err != nil {
			return err
		}

		logger.Info("subscriber", "name", ntr.Name, "notification", ntr.Spec.NotificationName)
		jsonParsed, err := gabs.ParseJSON(cron.DataFrom(ctx))
		if err != nil {
			panic(err)
		}

		v := jsonParsed.Path(ntr.Spec.WatchFieldPath).Data()

		switch v.(type) {
		case int, int32, int64:
			logger.Info("int value", "field", ntr.Spec.WatchFieldPath, "value", v)
		case float32, float64:
			logger.Info("float value", "field", ntr.Spec.WatchFieldPath, "value", v)
		case string:
			logger.Info("string value", "field", ntr.Spec.WatchFieldPath, "value", v)
		case map[string]interface{}:
			logger.Info("object value", "field", ntr.Spec.WatchFieldPath, "value", v)
		}
	}

	return nil
}

func parseSubscribers(o *tmaxiov1alpha1.Monitor) []types.NamespacedName {
	ret := []types.NamespacedName{}
	subscribers := strings.Split(o.Annotations["subscribers"], ",")
	if len(subscribers) == 1 && subscribers[0] == "" {
		return ret
	}
	for _, subscriber := range subscribers {
		tokens := strings.Split(subscriber, "/")
		ret = append(ret, types.NamespacedName{Namespace: tokens[0], Name: tokens[1]})
	}
	return ret
}

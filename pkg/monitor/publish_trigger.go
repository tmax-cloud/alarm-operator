package monitor

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
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

	logger := t.Logger.WithName("PublishTrigger").WithValues("monitor", t.Target.Name)

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
		if err := t.Client.Get(context.Background(), subscriber, ntr); err != nil {
			logger.Error(err, "")
			return err
		}

		nt := &tmaxiov1alpha1.Notification{}
		ntNamespaceName := types.NamespacedName{Namespace: subscriber.Namespace, Name: ntr.Spec.NotificationName}
		if err := t.Client.Get(context.Background(), ntNamespaceName, nt); err != nil {
			logger.Error(err, "")
			return err
		}

		if nt.Status.EndPoint == "" {
			err := fmt.Errorf("notification's endpoint not prepared")
			logger.Error(err, "")
			return err
		}

		logger.Info("subscriber", "name", ntr.Name, "notification", ntr.Spec.NotificationName)
		jsonParsed, err := gabs.ParseJSON(cron.DataFrom(ctx))
		if err != nil {
			return err
		}

		v := jsonParsed.Path(ntr.Spec.WatchFieldPath).Data()

		logger.Info("eval condition", "op1", v, "op2", ntr.Spec.Operand, "op", ntr.Spec.Op)
		if eval(v, ntr.Spec.Operand, ntr.Spec.Op) {
			logger.Info("Matched condition")
			postEndpoint(nt.Status.EndPoint)
		} else {
			logger.Info("Unmatched condition")
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

func eval(op1 interface{}, op2 string, op string) bool {
	switch op {
	case "gt", "<":
		switch op1 := op1.(type) {
		case int:
			operand, _ := strconv.Atoi(op2)
			if op1 > operand {
				return true
			}
		case float64:
			operand, _ := strconv.Atoi(op2)
			if op1 > float64(operand) {
				return true
			}
		case string:
			if op1 > op2 {
				return true
			}
		}
	case "gte", "<=":
		switch op1 := op1.(type) {
		case int:
			operand, _ := strconv.Atoi(op2)
			if op1 >= operand {
				return true
			}
		case float64:
			operand, _ := strconv.Atoi(op2)
			if op1 >= float64(operand) {
				return true
			}
		case string:
			if op1 >= op2 {
				return true
			}
		}
	case "eq", "==":
		switch op1 := op1.(type) {
		case int:
			operand, _ := strconv.Atoi(op2)
			if op1 == operand {
				return true
			}
		case float64:
			operand, _ := strconv.Atoi(op2)
			if op1 == float64(operand) {
				return true
			}
		case string:
			if op1 == op2 {
				return true
			}
		}
	case "lte", ">=":
		switch op1 := op1.(type) {
		case int:
			operand, _ := strconv.Atoi(op2)
			if op1 <= operand {
				return true
			}
		case float64:
			operand, _ := strconv.Atoi(op2)
			if op1 <= float64(operand) {
				return true
			}
		case string:
			if op1 <= op2 {
				return true
			}
		}
	case "lt", ">":
		switch op1 := op1.(type) {
		case int:
			operand, _ := strconv.Atoi(op2)
			if op1 < operand {
				return true
			}
		case float64:
			operand, _ := strconv.Atoi(op2)
			if op1 < float64(operand) {
				return true
			}
		case string:
			if op1 < op2 {
				return true
			}
		}
	}

	return false
}

func postEndpoint(url string) error {
	_, err := http.Post(url, "application/json", bytes.NewBuffer([]byte("")))
	if err != nil {
		return err
	}

	return nil
}

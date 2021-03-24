package monitor

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tmaxiov1alpha1 "github.com/tmax-cloud/alarm-operator/api/v1alpha1"
	"github.com/tmax-cloud/alarm-operator/pkg/cron"
)

type MonitorUpdater struct {
	client.Client
	target types.NamespacedName
	logger logr.Logger
}

func (r *MonitorUpdater) HandleFunc(next cron.TaskFunc) cron.TaskFunc {
	return func(ctx context.Context) error {
		r.handle(ctx)
		return next(ctx)
	}
}

func (r *MonitorUpdater) Handle(ctx context.Context) error {
	return r.handle(ctx)
}

func (r *MonitorUpdater) handle(ctx context.Context) error {

	logger := r.logger.WithName("MonitorUpdater")

	o := &tmaxiov1alpha1.Monitor{}
	err := r.Client.Get(ctx, r.target, o)
	if err != nil {
		return err
	}

	result := tmaxiov1alpha1.MonitorResult{}
	req, err := http.NewRequest("GET", o.Spec.URL, bytes.NewBuffer([]byte(o.Spec.Body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	result.LastTime = time.Now().Format(time.RFC3339)
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		result.Status = "Success"
		dat, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		result.Value = string(dat)
		ctx = cron.WithData(ctx, dat)
	} else {
		result.Status = "Fail"
	}

	o.Status.Result = result
	logger.Info("result", "status", result.Status, "value", result.Value, "time", result.LastTime)

	err = r.Client.Status().Update(context.Background(), o)
	if err != nil {
		return err
	}

	return nil
}

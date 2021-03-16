package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/tmax-cloud/alarm-operator/pkg/notification"
	"github.com/tmax-cloud/alarm-operator/pkg/notification/datasource"
	"github.com/tmax-cloud/alarm-operator/pkg/notifier/background"
	"github.com/tmax-cloud/alarm-operator/pkg/notifier/handler"
	"github.com/tmax-cloud/alarm-operator/pkg/notifier/job"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func main() {

	redisURL, _ := url.Parse(os.Getenv("REDIS_URL"))
	redisPort, _ := strconv.Atoi(strings.TrimPrefix(redisURL.Path, "/"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger = zap.NewExample().Sugar()
	defer logger.Sync()

	jobCh := background.NewWorker(ctx, logger)

	ds, err := datasource.NewRedisDataSource(redisURL.Host, "", redisPort)
	if err != nil {
		panic(err)
	}
	r := notification.NewNotificationRegistry(ds)
	q := notification.NewNotificationQueue(ds)

	go func() {
		for {
			// FIXME: Do not polling.
			noti, err := q.Dequeue()
			if err != nil {
				time.Sleep(time.Second)
				continue
			}

			jobCh <- job.NewNotificationJob(noti)
		}
	}()

	h := handler.New(ctx, r, q, logger)
	router := mux.NewRouter()
	router.Handle("/", h)

	s := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	logger.Info("Listening on :8080")
	logger.Fatal(s.ListenAndServe())
}

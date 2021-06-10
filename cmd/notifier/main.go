package main

import (
	"context"
	"flag"
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
	var port int
	flag.IntVar(&port, "port", 8080, "port number")
	flag.Parse()

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

	router := mux.NewRouter()
	router.Handle("/internal/notification/{namespace}/{id}", handler.NewRegistryHandler(ctx, r, logger)).Methods("POST")
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("I'm fine"))
	})
	router.Handle("/", handler.NewNotificationHandler(ctx, r, q, logger)).Methods("POST")

	s := &http.Server{
		Addr:           ":" + strconv.Itoa(port),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	logger.Info("Listening on ", port)
	logger.Fatal(s.ListenAndServe())
}

package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/tmax-cloud/alarm-operator/pkg/notification"
	"go.uber.org/zap"
)

type notificationHandler struct {
	registry *notification.NotificationRegistry
	queue    *notification.NotificationQueue
	logger   *zap.SugaredLogger
}

func New(ctx context.Context, registry *notification.NotificationRegistry, queue *notification.NotificationQueue, logger *zap.SugaredLogger) http.Handler {
	return &notificationHandler{
		registry: registry,
		queue:    queue,
		logger:   logger,
	}
}

func (h *notificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	h.logger.Infow("handler", "URL", r.URL, "Host", r.Host, "Hostname", r.URL.Hostname())

	id := extractIdFromHost(strings.Split(r.Host, ":")[0])
	notification, err := h.registry.Fetch(id)
	if err != nil {
		msg := fmt.Sprintf("Failed to fetch registry(id: %s): %s", id, err.Error())
		h.logger.Error(msg)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	h.logger.Infow("handler", "Host", r.Host, "extracted", id, "notification", notification)

	err = h.queue.Enqueue(notification)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Notification: %s reserved.", id)))
}

// extractIdFromHost extract XXXX from XXXX.127.0.0.1.nip.io
func extractIdFromHost(hostIn string) string {
	return strings.Split(hostIn, ".")[0]
}

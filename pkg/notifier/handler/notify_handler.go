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
	ctx      context.Context
	registry *notification.NotificationRegistry
	queue    *notification.NotificationQueue
	logger   *zap.SugaredLogger
}

func NewNotificationHandler(ctx context.Context, registry *notification.NotificationRegistry, queue *notification.NotificationQueue, logger *zap.SugaredLogger) http.Handler {
	return &notificationHandler{
		ctx:      ctx,
		registry: registry,
		queue:    queue,
		logger:   logger,
	}
}

func (h *notificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	authkey := r.Header.Get("AuthKey")
	h.logger.Infow("handler", "URL", r.URL, "Auth", authkey)

	id := extractIdFromHost(r.Host)

	key, noti, err := h.registry.Fetch(id)
	if err != nil {
		msg := fmt.Sprintf("Failed to fetch registry(id: %s): %s", id, err.Error())
		h.logger.Error(msg)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	if key != authkey {
		http.Error(w, "authoriation key not match", http.StatusNotFound)
		return
	}
	h.logger.Infow("handler", "Host", r.Host, "extracted", id, "notification", noti)

	err = h.queue.Enqueue(noti)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf("Notification: %s reserved.\n", id)))
}

// extractIdFromHost extract XXXX from XXXX.127.0.0.1.nip.io
func extractIdFromHost(hostIn string) string {
	id := strings.Split(hostIn, ".")[0]
	return id
}
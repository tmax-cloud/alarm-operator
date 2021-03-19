package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tmax-cloud/alarm-operator/pkg/notification"
	"go.uber.org/zap"
)

type registryHandler struct {
	registry *notification.NotificationRegistry
	logger   *zap.SugaredLogger
}

func NewRegistryHandler(ctx context.Context, registry *notification.NotificationRegistry, logger *zap.SugaredLogger) http.Handler {
	return &registryHandler{
		registry: registry,
		logger:   logger,
	}
}

func (h *registryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	h.logger.Infow("handler", "URL", r.URL, "Host", r.Host, "Hostname", r.URL.Hostname())

	id := mux.Vars(r)["id"]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}

	var noti notification.Notification
	switch strings.ToLower(r.URL.Query().Get("type")) {
	case "email":
		var n notification.MailNotification
		if err := json.Unmarshal(body, &n); err != nil {
			http.Error(w, "Failed to unmarshal body", http.StatusInternalServerError)
			return
		}
		noti = n
	case "webhook":
		var n notification.WebhookNotification
		if err := json.Unmarshal(body, &n); err != nil {
			http.Error(w, "Failed to unmarshal body", http.StatusInternalServerError)
			return
		}
		noti = n
	case "slack":
		var n notification.SlackNotification
		if err := json.Unmarshal(body, &n); err != nil {
			http.Error(w, "Failed to unmarshal body", http.StatusInternalServerError)
			return
		}
		noti = n
	default:
		http.Error(w, "Unknown notification type", http.StatusBadGateway)
		return
	}

	err = h.registry.Register(id, noti)
	if err != nil {
		msg := fmt.Sprintf("Failed to fetch registry(id: %s): %s", id, err.Error())
		h.logger.Error(msg)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Notification reserved"))
}

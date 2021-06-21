package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tmax-cloud/alarm-operator/pkg/notification"
	"github.com/tmax-cloud/alarm-operator/pkg/notification/ingress"
	"go.uber.org/zap"
)

type registryHandler struct {
	ctx      context.Context
	registry *notification.NotificationRegistry
	logger   *zap.SugaredLogger
}

func NewRegistryHandler(ctx context.Context, registry *notification.NotificationRegistry, logger *zap.SugaredLogger) http.Handler {
	return &registryHandler{
		ctx:      ctx,
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

	apikey, _, err := h.registry.Fetch(id)
	if err != nil {
		h.logger.Error(err)
	} else if apikey == "" {
		h.logger.Info("api key not found: ", id)
		if apikey, err = generateApiKey(); err != nil {
			msg := fmt.Sprintf("failed to generate apikey for %s", id)
			h.logger.Error(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
	}

	h.logger.Infow("new notification", "id", id, "apikey", apikey)
	err = h.registry.Register(id, apikey, noti)
	if err != nil {
		msg := fmt.Sprintf("Failed to fetch registry(id: %s): %s", id, err.Error())
		h.logger.Error(msg)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	go ingress.GetIngress(id)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(apikey))
}

func generateApiKey() (string, error) {
	bytes := make([]byte, 13)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

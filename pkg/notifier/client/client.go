package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tmax-cloud/alarm-operator/pkg/notification"
)

type Notifier struct {
	URL    string
	Client *http.Client
}

func New(u string, transport http.RoundTripper) *Notifier {

	return &Notifier{
		URL: u,
		Client: &http.Client{
			Transport: transport,
		},
	}
}

// Register regist notification
func (c *Notifier) Register(id, notiType string, noti notification.Notification) (*http.Response, error) {

	var payload []byte
	var err error

	if payload, err = json.Marshal(noti); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/internal/notification/%s?type=%s", c.URL, id, notiType)
	return http.Post(endpoint, "application/json", bytes.NewBuffer(payload))
}

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
func (c *Notifier) Register(id, notiType string, noti notification.Notification) error {

	var payload []byte
	var err error

	if payload, err = json.Marshal(noti); err != nil {
		return err
	}

	u := fmt.Sprintf("%s/internal/notification/%s?type=%s", c.URL, id, notiType)
	res, err := http.Post(u, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	dat, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	fmt.Println(string(dat))

	return nil
}

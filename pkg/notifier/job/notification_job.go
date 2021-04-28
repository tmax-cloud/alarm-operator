package job

import (
	"bytes"
	"crypto/tls"
	"encoding/json"

	"net/http"

	"github.com/tmax-cloud/alarm-operator/pkg/notification"
	"github.com/tmax-cloud/alarm-operator/pkg/notifier/background"
	"gopkg.in/gomail.v2"
)

type MailNotificationJob struct {
	noti notification.MailNotification
}

type WebhookNotificationJob struct {
	noti notification.WebhookNotification
}

type SlackNotificationJob struct {
	noti notification.SlackNotification
}

func NewNotificationJob(noti notification.Notification) background.Job {
	switch noti.(type) {
	case notification.MailNotification:
		return &MailNotificationJob{noti.(notification.MailNotification)}
	case notification.WebhookNotification:
		return &WebhookNotificationJob{noti.(notification.WebhookNotification)}
	case notification.SlackNotification:
		return &SlackNotificationJob{noti.(notification.SlackNotification)}
	}
	return nil
}

func (n *MailNotificationJob) Execute(job interface{}) error {
	m := gomail.NewMessage()
	m.SetHeader("From", n.noti.From)
	m.SetHeader("To", n.noti.To)
	// m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", n.noti.Subject)
	m.SetBody("text/html", n.noti.Body)

	d := gomail.NewDialer(n.noti.Host, n.noti.Port, n.noti.Username, n.noti.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func (n *WebhookNotificationJob) Execute(job interface{}) error {
	// TODO:
	return nil
}

func (n *SlackNotificationJob) Execute(job interface{}) error {
	// TODO:
	url := n.noti.Url
	msg := n.noti.Message
	reqbody, _ := json.Marshal(msg)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqbody))

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

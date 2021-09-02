package job

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	slackMessage := n.noti.SlackMessage
	pbytes, _ := json.Marshal(slackMessage)
	buff := bytes.NewBuffer(pbytes)

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", buff)
	if err != nil {
		return err
	}

	fmt.Println(n.noti.Authorization)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", n.noti.Authorization)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return err
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		str := string(respBody)
		fmt.Println(str)
	}

	return nil
}

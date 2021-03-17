package notification

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type NotificationQueue struct {
	ds Queue
}

func NewNotificationQueue(dataSource Queue) *NotificationQueue {
	return &NotificationQueue{ds: dataSource}
}

func (r *NotificationQueue) Enqueue(noti Notification) error {

	payload, err := json.Marshal(noti)
	if err != nil {
		return err
	}

	switch noti.(type) {
	case MailNotification, *MailNotification:
		payload = []byte(strings.Join([]string{"mail", string(payload)}, ":"))
	case WebhookNotification, *WebhookNotification:
		payload = []byte(strings.Join([]string{"webhook", string(payload)}, ":"))
	case SlackNotification, *SlackNotification:
		payload = []byte(strings.Join([]string{"slack", string(payload)}, ":"))
	default:
		return fmt.Errorf(fmt.Sprintf("unsupported notification type: %s\n", reflect.TypeOf(noti)))
	}

	return r.ds.Push(payload)
}

func (r *NotificationQueue) Dequeue() (Notification, error) {

	data, err := r.ds.Pop()
	if err != nil {
		return nil, err
	}

	// FIXME: too bad extraction
	tokens := strings.Split(string(data), ":")
	notiType := tokens[0]
	noti := strings.Join(append([]string{}, tokens[1:]...), ":")

	// FIXME: dirty
	switch notiType {
	case "mail":
		var dat MailNotification
		err := json.Unmarshal([]byte(noti), &dat)
		if err != nil {
			return nil, err
		}
		return dat, nil
	case "webhook":
		var dat WebhookNotification
		err := json.Unmarshal([]byte(noti), &dat)
		if err != nil {
			return nil, err
		}
		return dat, nil
	case "slack":
		var dat SlackNotification
		err := json.Unmarshal([]byte(noti), &dat)
		if err != nil {
			return nil, err
		}
		return dat, nil
	}

	return nil, fmt.Errorf("unknown type")
}

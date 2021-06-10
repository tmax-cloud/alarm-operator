package notification

type Notification interface{}

type MailMessage struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type SMTPConnection struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type SMTPAccount struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type MailNotification struct {
	SMTPConnection
	SMTPAccount
	MailMessage
}

type WebhookNotification struct {
	Url     string `json:"url"`
	Message string `json:"message"`
}

type SlackRequestBody struct {
	Text string `json:"text"`
}

type SlackNotification struct {
	Authorization       string `json:"authorization"`
	SlackMessage
}

type SlackMessage struct {
	Channel             string `json:"channel"`
	Text                string `json:"text"`
}

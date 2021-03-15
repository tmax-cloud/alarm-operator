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

type SlackNotification struct {
	SenderAccountSecret string `json:"account"`
	Workspace           string `json:"workspace"`
	Channel             string `json:"channel"`
	Message             string `json:"message"`
}

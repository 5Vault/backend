package external

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
)

type EmailClient struct {
	host     string
	port     string
	user     string
	password string
	fromName string
}

func NewEmailClient() *EmailClient {
	return &EmailClient{
		host:     "smtp.gmail.com",
		port:     "587",
		user:     os.Getenv("GMAIL_USER"),
		password: os.Getenv("GMAIL_APP_PASSWORD"),
		fromName: "FiveVault",
	}
}

func (e *EmailClient) Enabled() bool {
	return e.user != "" && e.password != ""
}

func (e *EmailClient) Send(to, subject, htmlBody string) error {
	if !e.Enabled() {
		return nil // silently skip if not configured
	}

	from := fmt.Sprintf("%s <%s>", e.fromName, e.user)
	msg := []byte(
		"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"From: " + from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n\r\n" +
			htmlBody,
	)

	auth := smtp.PlainAuth("", e.user, e.password, e.host)
	return smtp.SendMail(e.host+":"+e.port, auth, e.user, []string{to}, msg)
}

// RenderAndSend renders a named template with data and sends the email.
func (e *EmailClient) RenderAndSend(to, subject, tmplName string, data map[string]any) error {
	tmpl, ok := emailTemplates[tmplName]
	if !ok {
		return fmt.Errorf("email template %q not found", tmplName)
	}
	t, err := template.New(tmplName).Parse(tmpl)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}
	return e.Send(to, subject, buf.String())
}

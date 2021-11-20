package mail

import (
	"crypto/tls"

	"github.com/pkg/errors"
	gomail "gopkg.in/mail.v2"
)

type MailAuthConfig struct {
	Port         int
	Host         string
	Password     string
	EmailAddress string
}

type Client struct {
	dialer *gomail.Dialer
}

func (c *Client) Send(recipient, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", c.dialer.Username)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	if err := c.dialer.DialAndSend(m); err != nil {
		return errors.Wrapf(err, "failed to send email with body: %q", body)
	}

	return nil
}

type option func(*Client)

func WithDialer(mac MailAuthConfig) option {
	// Settings for SMTP server
	d := gomail.NewDialer(mac.Host, mac.Port, mac.EmailAddress, mac.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return func(client *Client) {
		client.dialer = d
	}
}

func NewMailClient(options ...option) *Client {
	client := &Client{}
	for _, opt := range options {
		opt(client)
	}

	return client
}

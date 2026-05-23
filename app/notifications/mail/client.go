// Package mail encapsula o envio de e-mails transacionais via Resend.
package mail

import (
	"fmt"
	"os"

	"github.com/resend/resend-go/v3"
)

// Client envia e-mails usando a API do Resend.
type Client struct {
	resend *resend.Client
	from   string
}

// NewClient constrói um Client lendo RESEND_KEY e MAIL_FROM do ambiente.
func NewClient() (*Client, error) {
	key := os.Getenv("RESEND_KEY")
	from := os.Getenv("MAIL_FROM")
	if key == "" {
		return nil, fmt.Errorf("RESEND_KEY not set")
	}
	if from == "" {
		return nil, fmt.Errorf("MAIL_FROM not set")
	}
	return &Client{
		resend: resend.NewClient(key),
		from:   from,
	}, nil
}

// SendEmail envia um e-mail de texto simples ao destinatário.
func (c *Client) SendEmail(to, subject, body string) error {
	_, err := c.resend.Emails.Send(&resend.SendEmailRequest{
		From:    c.from,
		To:      []string{to},
		Subject: subject,
		Text:    body,
	})
	return err
}

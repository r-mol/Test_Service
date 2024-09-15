package mail

import (
	"gopkg.in/gomail.v2"
)

type Mailer struct {
	dialer *gomail.Dialer
}

func NewMailer(
	cfg *Config,
) *Mailer {
	dialer := gomail.NewDialer(cfg.SmtpAddress, cfg.SmtpPort, cfg.AuthorName, cfg.AuthorPwd)

	return &Mailer{
		dialer: dialer,
	}
}

func (m *Mailer) SendMail(
	From string,
	To string,
	Subject string,
	Body string,
) error {
	message := gomail.NewMessage()
	message.SetHeader("From", From)
	message.SetHeader("To", To)
	message.SetHeader("Subject", Subject)
	message.SetBody("text/plain", Body)

	return m.dialer.DialAndSend(message)
}

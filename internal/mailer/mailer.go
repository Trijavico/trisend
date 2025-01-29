package mailer

import (
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"regexp"
	"trisend/internal/config"
)

const gmailHost = "smtp.gmail.com"

var (
	InvalidEmail = errors.New("invalid email address")
)

type mailer struct {
	subject  string
	receiver string
	body     string
}

func NewMailer(subject, receiverEmail, msg string) *mailer {
	return &mailer{
		subject:  subject,
		receiver: receiverEmail,
		body:     msg,
	}
}

func (mailer *mailer) Send() error {
	to := []string{mailer.receiver}

	auth := smtp.PlainAuth("", config.SMTP_USER, config.SMTP_PASSWORD, gmailHost)

	headers := make(map[string]string)
	headers["From"] = config.SMTP_USER
	headers["To"] = mailer.receiver
	headers["Subject"] = mailer.subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var message string
	for key, value := range headers {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	message += "\r\n" + mailer.body

	smtpServer := net.JoinHostPort(gmailHost, "587")
	if err := smtp.SendMail(smtpServer, auth, config.SMTP_USER, to, []byte(message)); err != nil {
		return err
	}

	return nil
}

func IsValidEmail(email string) bool {
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	validator := regexp.MustCompile(regex)
	if validator.MatchString(email) {
		return true
	}

	return false
}

package email

import (
	"fmt"

	"github.com/exmonitor/exclient/database/spec/service"
	"github.com/pkg/errors"
	"gopkg.in/gomail.v2"
)

type EmailConfig struct {
	To          string
	Failed      bool
	ServiceInfo *service.Service

	EmailChan   chan *gomail.Message
	SMTPEnabled bool
}

func NewEmail(conf EmailConfig) (*Email, error) {
	if conf.To == "" {
		return nil, errors.Wrap(invalidConfigError, "conf.To cannot be empty")
	}
	if conf.ServiceInfo == nil {
		return nil, errors.Wrap(invalidConfigError, "conf.ServiceInfo cannot be nil")
	}

	if conf.SMTPEnabled && conf.EmailChan == nil {
		return nil, errors.Wrap(invalidConfigError, "conf.EmailChan cannot be nil")
	}

	newEmail := &Email{
		to:          conf.To,
		failed:      conf.Failed,
		serviceInfo: conf.ServiceInfo,
		emailChan:   conf.EmailChan,
		smtpEnabled: conf.SMTPEnabled,
	}

	return newEmail, nil
}

type Email struct {
	to          string
	failed      bool
	serviceInfo *service.Service

	emailChan   chan *gomail.Message
	smtpEnabled bool
}

func (e *Email) Send() {
	if !e.smtpEnabled {
		fmt.Printf("<< fake email sent to %s\n", e.to)
		return
	}

	// build email struct
	msg := e.buildEmail()

	// send email to email daemon via channel
	e.emailChan <- msg
}

func (e *Email) buildEmail() *gomail.Message {
	m := gomail.NewMessage()

	m.SetHeader("To", e.to)
	m.SetHeader("Subject", e.emailSubject())
	m.SetBody("text/html", e.emailBody())

	return m
}

// to remove the complexity of the channel from other packages
func BuildEmailChannel() chan *gomail.Message {
	return make(chan *gomail.Message)
}

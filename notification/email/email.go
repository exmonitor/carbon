package email

import (
	"fmt"
	"github.com/exmonitor/exclient/database/spec/service"
	"github.com/pkg/errors"
	"gopkg.in/gomail.v2"
)

type EmailConfig struct {
	Failed      bool
	FailedMsg   string
	To          string
	ServiceInfo *service.Service
	SMTPEnabled bool

	SMTPEmailChan chan *gomail.Message
}

func NewEmail(conf EmailConfig) (*Email, error) {
	if conf.To == "" {
		return nil, errors.Wrap(invalidConfigError, "conf.To cannot be empty")
	}
	if conf.ServiceInfo == nil {
		return nil, errors.Wrap(invalidConfigError, "conf.ServiceInfo cannot be nil")
	}
	if conf.SMTPEnabled && conf.SMTPEmailChan == nil {
		return nil, errors.Wrap(invalidConfigError, "conf.SMTPEmailChan cannot be nil")
	}

	newEmail := &Email{
		failed:      conf.Failed,
		failedMsg:   conf.FailedMsg,
		to:          conf.To,
		serviceInfo: conf.ServiceInfo,
		smtpEnabled: conf.SMTPEnabled,

		emailChan: conf.SMTPEmailChan,
	}

	return newEmail, nil
}

type Email struct {
	failed      bool
	failedMsg   string
	to          string
	serviceInfo *service.Service

	emailChan   chan *gomail.Message
	smtpEnabled bool
}

func (e *Email) Send() {
	// build email struct
	msg := e.buildEmail()

	if e.smtpEnabled {
		// send email to email daemon via channel
		e.emailChan <- msg
	} else {
		fmt.Printf("<< fake email sent to %s\n %s\n", e.to, e.emailBody())
		return
	}

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

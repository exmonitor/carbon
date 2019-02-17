package email

import (
	"crypto/tls"
	"github.com/cenkalti/backoff"
	"github.com/exmonitor/exlogger"
	"github.com/pkg/errors"
	"gopkg.in/gomail.v2"
	"time"
)

type SMTPConfig struct {
	Server   string
	Port     int
	Username string
	Password string
	SMTPFrom string
}

type DaemonConfig struct {
	SMTPConfig SMTPConfig

	EmailChan chan *gomail.Message
	Logger    *exlogger.Logger
}

func NewDaemon(conf DaemonConfig) (*Daemon, error) {
	if conf.SMTPConfig.Server == "" {
		return nil, errors.Wrap(invalidConfigError, "conf.SMTPConfig.Server must not be empty")
	}
	if conf.SMTPConfig.Port == 0 {
		return nil, errors.Wrap(invalidConfigError, "conf.SMTPConfig.Port must not be zero")
	}
	if conf.SMTPConfig.Username == "" {
		return nil, errors.Wrap(invalidConfigError, "conf.SMTPConfig.Username must not be empty")
	}
	if conf.SMTPConfig.Password == "" {
		return nil, errors.Wrap(invalidConfigError, "conf.SMTPConfig.Password must not be empty")
	}
	if conf.SMTPConfig.SMTPFrom == "" {
		return nil, errors.Wrap(invalidConfigError, "conf.SMTPConfig.SMTPFrom must not be empty")
	}
	if conf.Logger == nil {
		return nil, errors.Wrap(invalidConfigError, "conf.Logger must not be nil")
	}
	if conf.EmailChan == nil {
		return nil, errors.Wrap(invalidConfigError, "conf.SMTPEmailChan must not be nil")
	}

	newDaemon := &Daemon{
		smtpConfig: conf.SMTPConfig,
		emailChan:  conf.EmailChan,
		logger:     conf.Logger,
	}
	return newDaemon, nil
}

type Daemon struct {
	smtpConfig SMTPConfig

	emailChan chan *gomail.Message
	logger    *exlogger.Logger
}

func (d *Daemon) StartDaemon() {
	/*
		// CRAM md5 AUTH
		auth := smtp.CRAMMD5Auth(d.smtpConfig.Username, d.smtpConfig.Password)
		// skip tls as for now we use self signed TLS fro postfix
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		// prepare smtp Dialer
		smtpDialer := &gomail.Dialer{
			Host:      d.smtpConfig.Server,
			Port:      d.smtpConfig.Port,
			SSL:       true,
			TLSConfig: tlsConfig,
			Auth:      auth,
		}*/
	smtpDialer := gomail.NewPlainDialer(d.smtpConfig.Server, d.smtpConfig.Port, d.smtpConfig.Username, d.smtpConfig.Password)
	smtpDialer.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	// run daemon
	go d.runDaemon(smtpDialer)
}

// run email daemon, which will send any email sent to emalChan
// if there is no email sent in last 30 sec, connection to SMTP will close and reopen again when new email is sent via channel
func (d *Daemon) runDaemon(smtpDialer *gomail.Dialer) {
	d.logger.Log("started email daemon")

	var s gomail.SendCloser
	var err error
	open := false
	for {
		select {
		// wait for message
		case m, ok := <-d.emailChan:
			// if channel is closed lets exit whole routine
			if !ok {
				d.logger.Log("emailChan is closed, stopping email daemon")
				return
			}
			// if connection to SMTP server is closed, open it
			if !open {
				// prepare backoff
				o := func() error {
					  s, err = smtpDialer.Dial()
					  return err
				}
				// execute login via backoff
				err = backoff.Retry(o, NewEmailBackoff(d.logger))
				if err != nil {
					panic(errors.Wrapf(err, "failed to connect to SMTP server"))
				}

				open = true
			}

			// assign 'From' email address
			m.SetHeader("From", fromHeader(d.smtpConfig.SMTPFrom, emailName))
			m.SetHeader("Return-Path", d.smtpConfig.SMTPFrom)

			// try send email with backoff
			o := func() error {
				err = gomail.Send(s, m)
				return err
			}

			// execute send email via backoff
			err = backoff.Retry(o, NewEmailBackoff(d.logger))
			if err != nil {
				d.logger.LogError(err, "failed to send email to %s", m.GetHeader("To"))
			} else {
				d.logger.LogDebug("sent email to %s", m.GetHeader("To"))
			}

		// Close the connection to the SMTP server if no email was sent in
		// the last 30 seconds.
		case <-time.After(30 * time.Second):
			if open {
				// close connection
				o := func() error {
					err = s.Close()
					return err
				}

				// execute send email via backoff
				err = backoff.Retry(o, NewEmailBackoff(d.logger))
				if err != nil {
					panic(err)
				}
				open = false
			}
		}
	}
}

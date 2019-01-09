package email

import (
	"crypto/tls"
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
		return nil, errors.Wrap(invalidConfigError, "conf.EmailChan must not be nil")
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
	// prepare smtp Dialer
	smtpDialer := gomail.NewPlainDialer(d.smtpConfig.Server, d.smtpConfig.Port, d.smtpConfig.Username, d.smtpConfig.Password)
	smtpDialer.SSL = true
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
				if s, err = smtpDialer.Dial(); err != nil {
					panic(err)
				}
				open = true
			}

			// assign 'From' email address
			m.SetBody("From", buildFromHeader(d.smtpConfig.SMTPFrom, emailName))

			// try send email
			if err := gomail.Send(s, m); err != nil {
				d.logger.LogError(err, "failed to send email")
			}

		// Close the connection to the SMTP server if no email was sent in
		// the last 30 seconds.
		case <-time.After(30 * time.Second):
			if open {
				// close connection
				err := s.Close()
				if err != nil {
					panic(err)
				}
				open = false
			}
		}
	}
}

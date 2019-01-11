package notification

import (
	"time"

	"github.com/exmonitor/exclient/database"
	dbnotification "github.com/exmonitor/exclient/database/spec/notification"
	"github.com/exmonitor/exclient/database/spec/service"
	"github.com/exmonitor/exlogger"
	"github.com/pkg/errors"

	"github.com/exmonitor/firefly/notification/email"
	"github.com/exmonitor/firefly/notification/phone"
	"github.com/exmonitor/firefly/notification/sms"
	"github.com/exmonitor/firefly/service/state"
	"gopkg.in/gomail.v2"
)

type Config struct {
	ServiceID                  int
	Failed                     bool
	FailedMsg                  string
	NotificationSentTimestamps map[int]time.Time
	NotificationChangeChannel  chan state.NotificationChange
	SMTPEnabled                bool
	SMTPEmailChan              chan *gomail.Message

	DBClient database.ClientInterface
	Logger   *exlogger.Logger
}

const (
	contactTypeEmail = "email"
	contactTypeSms   = "sms"
	contactTypePhone = "phone"
)

func New(conf Config) (*Service, error) {
	if conf.ServiceID <= 0 {
		return nil, errors.Wrap(invalidConfigError, "conf.DBCLient must be positive number")
	}
	if conf.DBClient == nil {
		return nil, errors.Wrap(invalidConfigError, "conf.DBCLient must not be nil")
	}
	if conf.Logger == nil {
		return nil, errors.Wrap(invalidConfigError, "conf.Logger must not be nil")
	}

	newService := &Service{
		checkId:                   conf.ServiceID,
		failed:                    conf.Failed,
		failedMsg:                 conf.FailedMsg,
		notificationSentTimestamp: conf.NotificationSentTimestamps,
		notificationChangeChannel: conf.NotificationChangeChannel,
		smtpEnabled:               conf.SMTPEnabled,
		smtpEmailChan:             conf.SMTPEmailChan,

		dbClient: conf.DBClient,
		logger:   conf.Logger,
	}

	return newService, nil
}

type Service struct {
	checkId                   int
	failed                    bool
	failedMsg                 string
	notificationSentTimestamp map[int]time.Time
	notificationChangeChannel chan state.NotificationChange
	smtpEnabled               bool
	smtpEmailChan             chan *gomail.Message

	dbClient database.ClientInterface
	logger   *exlogger.Logger
}

// for goroutine
func (s *Service) Run() {
	// fetch all user notification settings
	notificationSettings, err := s.dbClient.SQL_GetUsersNotificationSettings(s.checkId)
	if err != nil {
		s.logger.LogError(err, "failed to fetch user notification settings")
	}

	// get monitoring service details
	serviceInfo, err := s.dbClient.SQL_GetServiceDetails(s.checkId)
	if err != nil {
		s.logger.LogError(err, "failed to fetch service info")
	}

	for _, n := range notificationSettings {
		// check if we should resent notification
		if !s.canSentNotification(n) && s.failed {
			// notification was already sent and its still to early to resent
			continue
		}
		// execute notification
		s.executeNotification(serviceInfo, n)

	}
}

// func to to determine if notification should be sent
func (s *Service) canSentNotification(notificationSettings *dbnotification.UserNotificationSettings) bool {
	if notifTimestamp, ok := s.notificationSentTimestamp[notificationSettings.ID]; ok {
		// there is already record so this means notification was at sent at least once
		// let check if its time to resent
		if notificationSettings.ResentAfterMin == 1 {
			// notification settings 0 means dont resent notification ever
			return false
		}
		resentAfter := time.Duration(notificationSettings.ResentAfterMin) * time.Minute

		// checking if resent interval passed since last notification
		if time.Now().After(notifTimestamp.Add(resentAfter)) {
			nc := state.NotificationChange{
				ServiceID:      s.checkId,
				NotificationID: notificationSettings.ID,
			}
			// we dont need to save notification sent timestamp for OK notifications, as the records will be removed from DB anyway
			if s.failed {
				s.notificationChangeChannel <- nc
			}
			// sent notification
			s.logger.Log("resending notification for serviceID %d, notificationID %d after %.0fm", s.checkId, notificationSettings.ID, resentAfter.Minutes())
			return true
		} else {
			// interval for resending has not elapsed, dont sent notification
			return false
		}
	} else {
		// there is no record in notificationSentTimeStamp for this notify id, so should sent first notification
		nc := state.NotificationChange{
			ServiceID:      s.checkId,
			NotificationID: notificationSettings.ID,
		}
		// we dont need to save notification sent timestamp for OK notifications, as the records will be removed from DB anyway
		if s.failed {
			s.notificationChangeChannel <- nc
		}
		// sent notification
		s.logger.Log("send first notification for serviceID %d, notificationID %d", s.checkId, notificationSettings.ID)
		return true
	}

}

func (s *Service) executeNotification(serviceInfo *service.Service, n *dbnotification.UserNotificationSettings) {
	switch n.Type {
	case contactTypeEmail:
		// prepare email config
		emailConfig := email.EmailConfig{
			To:            n.Target,
			Failed:        s.failed,
			FailedMsg:     s.failedMsg,
			ServiceInfo:   serviceInfo,
			SMTPEnabled:   s.smtpEnabled,
			SMTPEmailChan: s.smtpEmailChan,
		}

		emailSender, err := email.NewEmail(emailConfig)
		if err != nil {
			s.logger.LogError(err, "failed to prepare Email to %s for check id %d", n.Target, s.checkId)
		}
		// send email
		emailSender.Send()
		break
	case contactTypeSms:
		msg := SMSTemplate(s.failed, serviceInfo)
		err := sms.Send(n.Target, msg)
		if err != nil {
			s.logger.LogError(err, "failed to send SMS to %s for check id %d", n.Target, s.checkId)
		}
		break
	case contactTypePhone:
		msg := CallTemplate(s.failed, serviceInfo)
		err := phone.Call(n.Target, msg)
		if err != nil {
			s.logger.LogError(err, "failed to call to %s for check id %d", n.Target, s.checkId)
		}
	default:
		s.logger.LogError(unknownContactTypeError, "contact type %s not recognized", n.Type)
	}
}

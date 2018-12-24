package notification

import (
	"github.com/exmonitor/firefly/database"
	"github.com/exmonitor/firefly/log"
	"github.com/exmonitor/firefly/notification/email"
	"github.com/exmonitor/firefly/notification/phone"
	"github.com/exmonitor/firefly/notification/sms"
	"github.com/pkg/errors"
)

type Config struct {
	ServiceID int
	Failed    bool
	DBClient  database.ClientInterface
	Logger    *log.Logger
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

	newService := &Service{
		checkId: conf.ServiceID,
		failed:  conf.Failed,

		dbClient: conf.DBClient,
	}

	return newService, nil
}

type Service struct {
	checkId  int
	failed   bool
	dbClient database.ClientInterface
	logger   *log.Logger
}

// for goroutine
func (s *Service) Run() {
	// get monitoring service details
	serviceInfo, err := s.dbClient.SQL_GetServiceDetails(s.checkId)
	if err != nil {
		s.logger.LogError(err, "failed to fetch service info")
	}
	// fetch all user notification settings
	notificationSettings, err := s.dbClient.SQL_GetUsersNotificationSettings(s.checkId)
	if err != nil {
		s.logger.LogError(err, "failed to fetch user notification settings")
	}

	for _, n := range notificationSettings {
		switch n.Type {
		case contactTypeEmail:
			msg := EmailTemplate(s.failed, serviceInfo)
			err := email.Send(n.Target, msg)
			if err != nil {
				s.logger.LogError(err, "failed to send Email to %s for check id %d", n.Target, s.checkId)
			}
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
		}
	}

}

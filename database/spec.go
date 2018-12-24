package database

import (
	"time"

	"github.com/exmonitor/firefly/database/spec/status"
	"github.com/exmonitor/firefly/database/spec/notification"
	"github.com/exmonitor/firefly/database/spec/service"
)

type ClientInterface interface {
	// elastic queries
	ES_GetServiceStateResults(from time.Time, to time.Time, interval int) ([]*status.Status, error)

	// maria queries
	SQL_GetServiceDetails(checkID int) (*service.Service, error)
	SQL_GetUsersNotificationSettings(checkId int) ([]*notification.UserNotificationSettings, error)
	SQL_GetIntervals() ([]int,error)
}


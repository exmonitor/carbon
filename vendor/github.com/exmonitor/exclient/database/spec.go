package database

import (
	"time"

	"github.com/exmonitor/exclient/database/spec/notification"
	"github.com/exmonitor/exclient/database/spec/service"
	"github.com/exmonitor/exclient/database/spec/status"
	"github.com/olivere/elastic"
)

type ClientInterface interface {
	// client specific
	Close()

	// elastic queries
	ES_GetFailedServices(from time.Time, to time.Time, interval int) ([]*status.ServiceStatus, error)
	ES_GetServicesStatus(from time.Time, to time.Time, elasticQuery ...elastic.Query) ([]*status.ServiceStatus, error)
	ES_SaveServiceStatus(s *status.ServiceStatus) error
	ES_DeleteServicesStatus(from time.Time, to time.Time) error
	ES_GetAggregatedServiceStatusByID(from time.Time, to time.Time, serviceID int) (*status.AgregatedServiceStatus, error)
	ES_SaveAggregatedServiceStatus(s *status.AgregatedServiceStatus) error

	// maria queries
	SQL_GetServices(intervalSec int) ([]*service.Service, error)
	SQL_GetServiceDetails(serviceID int) (*service.Service, error)
	SQL_GetUsersNotificationSettings(serviceID int) ([]*notification.UserNotificationSettings, error)
	SQL_GetIntervals() ([]int, error)
}

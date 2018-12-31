package multi

import (
	"github.com/exmonitor/chronos"
	"github.com/pkg/errors"

	"github.com/exmonitor/exclient/database/spec/notification"
	"github.com/exmonitor/exclient/database/spec/service"
)

// ********************************************
// MARIA DB
//----------------------------------------------

// intervals table
/*
| intervalSec | CREATE TABLE `intervalSec` (
  `id_interval` int(5) NOT NULL AUTO_INCREMENT,
  `value` int(5) NOT NULL,
  PRIMARY KEY (`id_interval`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 |

*/
func (c *Client) SQL_GetIntervals() ([]int, error) {
	t := chronos.New()
	q := "SELECT " +
		"id_interval, " +
		"value " +
		"FROM " +
		"intervalSec"
	// create sql query
	rows, err := c.sqlClient.Query(q)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute SQL_GetIntervals")
	}

	var intervals []int
	// read result
	for rows.Next() {
		var id, value int
		err := rows.Scan(&id, &value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan values in SQL_GetIntervals")
		}
		intervals = append(intervals, value)
	}

	c.logger.LogDebug("fetched %d intervals from SQL", len(intervals))
	t.Finish()
	if c.timeProfiling {
		c.logger.LogDebug("TIME_PROFILING: executed SQL_GetIntervals in %sms", t.StringMilisec())
	}
	return intervals, nil
}

/*
| notification | CREATE TABLE `notification` (
  `id_notification` int(5) NOT NULL AUTO_INCREMENT,
  `type` varchar(30) NOT NULL,
  `target` varchar(30) NOT NULL,
  `fk_users` int(11) NOT NULL,
  `fk_settings` int(11) NOT NULL,
  PRIMARY KEY (`id_notification`),
  KEY `fk_users` (`fk_users`),
  KEY `fk_settings` (`fk_settings`),
  CONSTRAINT `notify_settings` FOREIGN KEY (`fk_settings`) REFERENCES `notify_settings` (`id_settings`),
  CONSTRAINT `users` FOREIGN KEY (`fk_users`) REFERENCES `users` (`id_users`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 |
*/
func (c *Client) SQL_GetUsersNotificationSettings(serviceID int) ([]*notification.UserNotificationSettings, error) {
	t := chronos.New()
	q := "SELECT " +
		"notification.id_notification, " +
		"notification.type, " +
		"notification.target, " +
		"notify_settings.intervalMin " +
		"FROM " +
		"services " +
		"JOIN hosts ON fk_service_hosts=id_hosts " +
		"JOIN users ON hosts.fk_customer=users.fk_customer " +
		"JOIN notification ON id_users=fk_users " +
		"JOIN notify_settings ON fk_settings=id_settings " +
		"WHERE services.id_services = ?;"
	// prepare sql query
	query, err := c.sqlClient.Prepare(q)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare query SQL_GetUsersNotificationSettings")
	}
	// execute sql query
	rows, err := query.Query(serviceID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute SQL_GetUsersNotificationSettings")
	}
	var notifications []*notification.UserNotificationSettings

	// read result
	for rows.Next() {
		var target, notificationType string
		var id, resentAfterMin int
		// scan rows
		err := rows.Scan(&id, &notificationType, &target, &resentAfterMin)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan values in SQL_GetUsersNotificationSettings")
		}
		// init UserNotificationSettings struct
		n := &notification.UserNotificationSettings{
			ID:             id,
			Target:         target,
			Type:           notificationType,
			ResentAfterMin: resentAfterMin,
		}
		notifications = append(notifications, n)
	}

	t.Finish()
	if c.timeProfiling {
		c.logger.LogDebug("TIME_PROFILING: executed SQL_GetUsersNotificationSettings:ID:%d in %sms", serviceID, t.StringMilisec())
	}
	return notifications, nil
}

func (c *Client) SQL_GetServices(interval int) ([]*service.Service, error) {
	t := chronos.New()
	q := "SELECT " +
		"services.id_services, " +
		"services.fail_treshold, " +
		"intervalSec.value, " +
		"service_metadata.metadata, " +
		"services.fk_service_type, " +
		"hosts.dns_or_ip, " +
		"hosts.extra_info, " +
		"location.name " +
		"FROM " +
		"services " +
		"JOIN intervalSec on fk_interval=id_interval " +
		"JOIN service_metadata ON fk_service_metadata=id_service_metadata " +
		"JOIN hosts ON fk_service_hosts=id_hosts " +
		"JOIN location ON fk_location=id_location " +
		"WHERE intervalSec.value=?;"
	// prepare sql query
	query, err := c.sqlClient.Prepare(q)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare query SQL_GetServices")
	}
	// execute sql query
	rows, err := query.Query(interval)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute SQL_GetServices")
	}
	var services []*service.Service

	// read result
	for rows.Next() {
		var serviceID, failThreshold, intervalSec, serviceType int
		var serviceMetadata, hostTarget, hostName, location string
		// scan rows
		err := rows.Scan(&serviceID, &failThreshold, &intervalSec, &serviceMetadata, &serviceType, &hostTarget, &hostName, &location)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan values in SQL_GetServices")
		}
		// init service struct
		s := &service.Service{
			ID:            serviceID,
			FailThreshold: failThreshold,
			Metadata:      serviceMetadata,
			Type:          serviceType,
			Target:        hostTarget,
			Host:          hostName,
			Interval:      intervalSec,
		}
		services = append(services, s)
	}

	t.Finish()
	if c.timeProfiling {
		c.logger.LogDebug("TIME_PROFILING: executed SQL_GetServices:%d in %sms", interval, t.StringMilisec())
	}
	return services, nil
}

func (c *Client) SQL_GetServiceDetails(serviceID int) (*service.Service, error) {
	t := chronos.New()
	q := "SELECT " +
		"services.id_services, " +
		"services.fail_treshold, " +
		"intervalSec.value, " +
		"service_metadata.metadata, " +
		"services.fk_service_type, " +
		"hosts.dns_or_ip, " +
		"hosts.extra_info, " +
		"location.name " +
		"FROM " +
		"services " +
		"JOIN intervalSec on fk_interval=id_interval " +
		"JOIN service_metadata ON fk_service_metadata=id_service_metadata " +
		"JOIN hosts ON fk_service_hosts=id_hosts " +
		"JOIN location ON fk_location=id_location " +
		"WHERE services.id_services=?;"
	// prepare sql query
	query, err := c.sqlClient.Prepare(q)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare query SQL_GetServiceDetails")
	}
	// execute sql query
	rows, err := query.Query(serviceID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute SQL_GetServiceDetails")
	}

	var s *service.Service
	if rows.Next() {
		// read result
		var failThreshold, intervalSec, serviceType int
		var serviceMetadata, hostTarget, hostName, location string
		// scan rows
		err = rows.Scan(&serviceID, &failThreshold, &intervalSec, &serviceMetadata, &serviceType, &hostTarget, &hostName, &location)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan values in SQL_GetServiceDetails")
		}
		// init service struct
		s = &service.Service{
			ID:            serviceID,
			FailThreshold: failThreshold,
			Metadata:      serviceMetadata,
			Type:          serviceType,
			Target:        hostTarget,
			Host:          hostName,
			Interval:      intervalSec,
		}
	} else {
		return nil, errors.Wrapf(executionFailedError, "failed to fetch service ID %d, no results found in db", serviceID)
	}

	t.Finish()
	if c.timeProfiling {
		c.logger.LogDebug("TIME_PROFILING: executed SQL_GetServiceDetails:ID:%d in %sms", serviceID, t.StringMilisec())
	}

	return s, nil
}

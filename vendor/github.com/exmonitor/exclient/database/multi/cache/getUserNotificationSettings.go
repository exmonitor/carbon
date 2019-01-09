package cache

import (
	"time"

	"github.com/exmonitor/exclient/database/spec/notification"
)

type SQL_GetUsersNotificationSetting struct {
	Cache map[int]SQL_GetUsersNotificationSetting_Record
}

type SQL_GetUsersNotificationSetting_Record struct {
	Age  time.Time
	Data []*notification.UserNotificationSettings
}

// check if cache is still valid
// returns false in case there is no cache or cache is already expired
func (s *SQL_GetUsersNotificationSetting) IsCacheValid(serviceID int, ttl time.Duration) bool {
	if r, ok := s.Cache[serviceID]; ok {
		if r.Age.IsZero() {
			// cache age is not set, cache is not valid
			return false
		} else {
			return time.Now().Before(r.Age.Add(ttl))
		}
	} else {
		// no cache for this record, so cache is not valid
		return false
	}
}

// get cached data
func (s *SQL_GetUsersNotificationSetting) GetData(serviceID int) []*notification.UserNotificationSettings {
	if d, ok := s.Cache[serviceID]; ok {
		return d.Data
	}
	// cached data not found
	return nil
}

// save data to cache
func (s *SQL_GetUsersNotificationSetting) CacheData(serviceID int, d []*notification.UserNotificationSettings) {
	r := SQL_GetUsersNotificationSetting_Record{
		Age:  time.Now(),
		Data: d,
	}
	s.Cache[serviceID] = r
}

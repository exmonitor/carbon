package cache

import (
	"time"

	"github.com/exmonitor/exclient/database/spec/service"
)

type SQL_GetServiceDetails struct {
	Cache map[int]SQL_GetServiceDetails_Record
}

type SQL_GetServiceDetails_Record struct {
	Age  time.Time
	Data *service.Service
}

// check if cache is still valid
// returns false in case there is no cache or cache is already expired
func (s *SQL_GetServiceDetails) IsCacheValid(serviceID int, ttl time.Duration) bool {
	if r, ok := s.Cache[serviceID]; ok {
		if r.Age.IsZero() {
			// cache age is not set, cache is not balid
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
func (s *SQL_GetServiceDetails) GetData(serviceID int) *service.Service {
	if d, ok := s.Cache[serviceID]; ok {
		return d.Data
	}
	// cached data not found
	return nil
}

// save data to cache
func (s *SQL_GetServiceDetails) CacheData(serviceID int, d *service.Service) {
	r := SQL_GetServiceDetails_Record{
		Age:  time.Now(),
		Data: d,
	}
	s.Cache[serviceID] = r
}

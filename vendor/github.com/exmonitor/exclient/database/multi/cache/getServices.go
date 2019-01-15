package cache

import (
	"time"

	"github.com/exmonitor/exclient/database/spec/service"
	"sync"
)

type SQL_GetServices struct {
	Cache map[int]SQL_GetServices_Record
	sync.Mutex
}

type SQL_GetServices_Record struct {
	Age  time.Time
	Data []*service.Service
}

// check if cache is still valid
// returns false in case there is no cache or cache is already expired
func (s *SQL_GetServices) IsCacheValid(intervalSec int, ttl time.Duration) bool {
	if r, ok := s.Cache[intervalSec]; ok {
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

// fetch data from cache
func (s *SQL_GetServices) GetData(interval int) []*service.Service {
	if d, ok := s.Cache[interval]; ok {
		return d.Data
	}
	// cached data not found
	return nil
}

// save data to cache
func (s *SQL_GetServices) CacheData(intervalSec int, d []*service.Service) {
	s.Lock()

	r := SQL_GetServices_Record{
		Age:  time.Now(),
		Data: d,
	}
	s.Cache[intervalSec] = r

	s.Unlock()
}

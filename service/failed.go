package service

import (
	"sync"
	"time"
)

type FailedService struct {
	Id            int
	FailCounter   int
	FailThreshold int
	LastFailedMsg string

	NotificationSentTimestamps map[int]time.Time
	sync.Mutex
}

// atomic save into map
func (f *FailedService) SaveNewTimeStamp(id int, t time.Time) {
	f.Lock()
	f.NotificationSentTimestamps[id] = t
	f.Unlock()
}

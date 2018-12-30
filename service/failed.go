package service

import "time"

type FailedService struct {
	Id            int
	FailCounter   int
	FailThreshold int
	LastFailedMsg string

	NotificationSentTimestamps map[int]time.Time
}

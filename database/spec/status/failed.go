package status

import "time"

type FailedService struct {
	Id            int
	FailCounter   int
	FailThreshold int
	LastFailedMsg string

	SentNotification     bool
	SentNotificationTime time.Time

	ResentEvery time.Duration
}

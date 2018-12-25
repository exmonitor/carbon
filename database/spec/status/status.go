package status

import "time"

type FailedStatus struct {
	Id            int
	ReqId         string
	FailThreshold int
	ResentEvery   time.Duration
	Duration      time.Duration
	Message       string
}

// find status in the status array
func FindStatus(s []*FailedStatus, id int) *FailedStatus {
	for _, status := range s {
		if status.Id == id {
			return status
		}
	}
	return nil
}

// check if status with specific id exists in the status array
func Exists(s []*FailedStatus, id int) bool {
	return FindStatus(s, id) != nil
}

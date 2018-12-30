package status

import "time"

type ServiceStatus struct {
	Id            int           `json:"id"`
	FailThreshold int           `json:"failThreshold"`
	Duration      time.Duration `json:"duration"`
	Message       string        `json:"message"`
	Result        bool          `json:"result"`
	ReqId         string        `json:"reqId"`
	// used for saving timestamp when it was inserted into ES DB
	InsertTimestamp time.Time `json:"@timestamp"`
}

// find status in the status array
func FindStatus(s []*ServiceStatus, id int) *ServiceStatus {
	for _, status := range s {
		if status.Id == id {
			return status
		}
	}
	return nil
}

// check if status with specific id exists in the status array
func Exists(s []*ServiceStatus, id int) bool {
	return FindStatus(s, id) != nil
}

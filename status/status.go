package status

import "time"

type Status struct {
	Id            int
	ReqId         string
	Result        bool
	FailThreshold int
	Duration      time.Duration
	Message       string
}

// return bool that decide if monitoring check failed
func (s *Status) Failed() bool {
	return !s.Result
}

// find status in the status array
func FindStatus(s []*Status, id int) *Status {
	for _, status := range s {
		if status.Id == id {
			return status
		}
	}
	return nil
}

// check if status with specific id exists in the status array
func StatusExists(s []*Status, id int) bool {
	return FindStatus(s, id) != nil
}

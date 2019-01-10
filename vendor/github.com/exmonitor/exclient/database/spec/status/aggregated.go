package status

import (
	"fmt"
	"time"
)

type AgregatedServiceStatus struct {
	Id            string        `json:"id"`
	ServiceID     int           `json:"service_id"`
	Interval      int           `json:"interval"`
	AvgDuration   time.Duration `json:"avg_duration"`
	Aggregated    int           `json:"aggregated"`
	Result        bool          `json:"result"`
	TimestampFrom time.Time     `json:"@timestamp_from"`
	TimestampTo   time.Time     `json:"@timestamp_to"`
}

func (a *AgregatedServiceStatus) String() string {
	if a == nil {
		return "[nil]"
	}
	return fmt.Sprintf("[id:%s, serviceID: %d, interval: %d, aggregated: %d, result: %t, from: %s, to: %s]", a.Id, a.ServiceID, a.Interval, a.Aggregated, a.Result, simpleTimeFormat(a.TimestampFrom), simpleTimeFormat(a.TimestampTo))
}

func simpleTimeFormat(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

package multi

import (
	"time"

	"github.com/exmonitor/firefly/database/spec/status"
)

// **************************************************
// ELASTIC SEARCH
///--------------------------------------------------
func (c *Client) ES_GetFailedServices(from time.Time, to time.Time, interval int) ([]*status.FailedStatus, error) {
	// just dummy record return

	return nil, nil
}

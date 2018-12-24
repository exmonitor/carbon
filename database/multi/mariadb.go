package multi

import (
	"github.com/exmonitor/firefly/database/spec/notification"
	"github.com/exmonitor/firefly/database/spec/service"
)

// ********************************************
// MARIA DB
//----------------------------------------------
func (c *Client) SQL_GetIntervals() ([]int, error) {
	// TODO
	return []int{30, 60, 120, 300}, nil
}

func (c *Client) SQL_GetUsersNotificationSettings(checkId int) ([]*notification.UserNotificationSettings, error) {

	return nil, nil
}

func (c *Client) SQL_GetServiceDetails(checkID int) (*service.Service, error) {

	return nil, nil
}

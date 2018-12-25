package dummydb

import (
	"time"

	"github.com/exmonitor/firefly/database"
	"github.com/exmonitor/firefly/database/spec/notification"
	"github.com/exmonitor/firefly/database/spec/service"
	"github.com/exmonitor/firefly/database/spec/status"
	"fmt"
)

type Config struct {
	// no config for dummydb needed
}

func DBDriverName() string {
	return "dummydb"
}

type Client struct {
	// implement client db interface
	database.ClientInterface
}

func GetClient(config Config) *Client {
	return &Client{}
}

var dummyDBStatusCounter = 0
var dummyDBStatusIncreaser = 1

// **************************************************
// ELASTIC SEARCH
///--------------------------------------------------
func (c *Client) ES_GetFailedServices(from time.Time, to time.Time, interval int) ([]*status.FailedStatus, error) {
	// just dummy record return
	var statusArray []*status.FailedStatus

	fmt.Printf("DummyDB| ES_GetFailedServices | statusCounter %d, statusIncreaser %d\n", dummyDBStatusCounter, dummyDBStatusIncreaser)

	// simulate change, we sent 10x  failed and than 10 non-failed for some services
	if  dummyDBStatusIncreaser > 0 {
		dummyDBStatusCounter += dummyDBStatusIncreaser
		if dummyDBStatusCounter > 25 {
			dummyDBStatusIncreaser = -1
		}
		if interval == 30 {

			status1 := &status.FailedStatus{
				Id:            1,
				Duration:      time.Second,
				Message:       "OK",
				FailThreshold: 5,
				ReqId:         "xxxx",
				ResentEvery:   time.Minute * 60,
			}
			status2 := &status.FailedStatus{
				Id:            2,
				Duration:      time.Second,
				Message:       "check tcp: connection time out",
				FailThreshold: 5,
				ReqId:         "xxxxsssss",
				ResentEvery:   time.Minute * 15,
			}

			statusArray = append(statusArray, status1)
			statusArray = append(statusArray, status2)
		} else if interval == 60 {
			status3 := &status.FailedStatus{
				Id:            3,
				Duration:      time.Second,
				Message:       "check tcp: connection refused",
				FailThreshold: 3,
				ReqId:         "xxxxzzzz",
				ResentEvery:   time.Minute * 5,
			}
			status4 := &status.FailedStatus{
				Id:            4,
				Duration:      time.Second,
				Message:       "check http: returned 503 status",
				FailThreshold: 5,
				ReqId:         "xxxxyyyy",
				ResentEvery:   time.Minute * 5,
			}

			statusArray = append(statusArray, status3)
			statusArray = append(statusArray, status4)
		}
	} else {
		dummyDBStatusCounter += dummyDBStatusIncreaser
		if dummyDBStatusCounter <= 0 {
			dummyDBStatusIncreaser = 1
		}
		if interval == 30 {

			status2 := &status.FailedStatus{
				Id:            2,
				Duration:      time.Second,
				Message:       "check tcp: connection time out",
				FailThreshold: 5,
				ReqId:         "xxxxsssss",
				ResentEvery:   time.Minute * 15,
			}

			statusArray = append(statusArray, status2)
		} else if interval == 60 {
			status3 := &status.FailedStatus{
				Id:            3,
				Duration:      time.Second,
				Message:       "check tcp: connection refused",
				FailThreshold: 3,
				ReqId:         "xxxxzzzz",
				ResentEvery:   time.Minute * 5,
			}

			statusArray = append(statusArray, status3)
		}
	}

	return statusArray, nil
}

// ********************************************
// MARIA DB
//----------------------------------------------
func (c *Client) SQL_GetIntervals() ([]int, error) {
	return []int{30, 60, 120}, nil
}

func (c *Client) SQL_GetUsersNotificationSettings(checkId int) ([]*notification.UserNotificationSettings, error) {
	var userNotifSettings []*notification.UserNotificationSettings

	if checkId == 1 {
		// user1 email
		user1Notif := &notification.UserNotificationSettings{
			Target: "jardaID1@seznam.cz",
			Type: "email",
		}

		userNotifSettings = append(userNotifSettings, user1Notif)
	} else if checkId == 2 {
		// user1 email
		user1Notif := &notification.UserNotificationSettings{
			Target: "jardaID2@seznam.cz",
			Type: "email",
		}
		// user3 email
		user2Notif := &notification.UserNotificationSettings{
			Target: "123456789ID2",
			Type: "sms",
		}

		userNotifSettings = append(userNotifSettings, user1Notif)
		userNotifSettings = append(userNotifSettings, user2Notif)
	} else if checkId == 3 {
		// user1 email
		user1Notif := &notification.UserNotificationSettings{
			Target: "TomosID3@seznam.cz",
			Type: "email",
		}

		userNotifSettings = append(userNotifSettings, user1Notif)

	} else if checkId == 4 {
		// user1 email
		user1Notif := &notification.UserNotificationSettings{
			Target: "456789854ID4",
			Type: "sms",
		}

		userNotifSettings = append(userNotifSettings, user1Notif)
	}

	return userNotifSettings, nil
}

func (c *Client) SQL_GetServiceDetails(checkID int) (*service.Service, error) {
	var  serviceDetail *service.Service
	if checkID == 1 {
		serviceDetail = &service.Service{
			ID:1,
			Host:"myServer1",
			Target:"web.myserver.com",
			Port:80,
			ServiceType:1,
			FailThreshold:5,
			Interval:30,
		}
	} else if checkID == 2 {
		serviceDetail = &service.Service{
			ID:2,
			Host:"myWeb1",
			Target:"webik.com",
			Port:443,
			ServiceType:1,
			FailThreshold:5,
			Interval:30,
		}

	} else if checkID == 3 {
		serviceDetail = &service.Service{
			ID:3,
			Host:"bigServer",
			Target:"seznam.com",
			Port:8080,
			ServiceType:1,
			FailThreshold:3,
			Interval:30,
		}

	} else if checkID == 4 {
		serviceDetail = &service.Service{
			ID:4,
			Host:"myICMPTestServer",
			Target:"google.com",
			Port:0,
			ServiceType:2,
			FailThreshold:3,
			Interval:30,
		}

	}


	return serviceDetail, nil
}

package dummydb

import (
	"fmt"
	"time"

	"github.com/exmonitor/exclient/database"
	"github.com/exmonitor/exclient/database/spec/notification"
	"github.com/exmonitor/exclient/database/spec/service"
	"github.com/exmonitor/exclient/database/spec/status"
	"github.com/olivere/elastic"
	"github.com/exmonitor/exlogger"
)

type Config struct {
	Logger *exlogger.Logger

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
	config.Logger.Log("using DUMMYDB driver")
	return &Client{}
}

func (c *Client) Close() {

}

var dummyDBStatusCounter = 0
var dummyDBStatusIncreaser = 1

const (
	timeLayout = "2006-01-02 15:04:05"
)

// **************************************************
// ELASTIC SEARCH
///--------------------------------------------------
func (c *Client) ES_GetFailedServices(from time.Time, to time.Time, interval int) ([]*status.ServiceStatus, error) {
	// just dummy record return
	var statusArray []*status.ServiceStatus

	fmt.Printf("DummyDB| ES_GetFailedServices | statusCounter %d, statusIncreaser %d\n", dummyDBStatusCounter, dummyDBStatusIncreaser)

	// simulate change, we sent 10x  failed and than 10 non-failed for some services
	if dummyDBStatusIncreaser > 0 {
		dummyDBStatusCounter += dummyDBStatusIncreaser
		if dummyDBStatusCounter > 25 {
			dummyDBStatusIncreaser = -1
		}
		if interval == 30 {

			status1 := &status.ServiceStatus{
				Id:            1,
				Duration:      time.Second,
				Message:       "OK",
				FailThreshold: 5,
				ReqId:         "xxxx",
			}
			status2 := &status.ServiceStatus{
				Id:            2,
				Duration:      time.Second,
				Message:       "check tcp: connection time out",
				FailThreshold: 5,
				ReqId:         "xxxxsssss",
			}

			statusArray = append(statusArray, status1)
			statusArray = append(statusArray, status2)
		} else if interval == 60 {
			status3 := &status.ServiceStatus{
				Id:            3,
				Duration:      time.Second,
				Message:       "check tcp: connection refused",
				FailThreshold: 3,
				ReqId:         "xxxxzzzz",
			}
			status4 := &status.ServiceStatus{
				Id:            4,
				Duration:      time.Second,
				Message:       "check http: returned 503 status",
				FailThreshold: 5,
				ReqId:         "xxxxyyyy",
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

			status2 := &status.ServiceStatus{
				Id:            2,
				Duration:      time.Second,
				Message:       "check tcp: connection time out",
				FailThreshold: 5,
				ReqId:         "xxxxsssss",
			}

			statusArray = append(statusArray, status2)
		} else if interval == 60 {
			status3 := &status.ServiceStatus{
				Id:            3,
				Duration:      time.Second,
				Message:       "check tcp: connection refused",
				FailThreshold: 3,
				ReqId:         "xxxxzzzz",
			}

			statusArray = append(statusArray, status3)
		}
	}

	return statusArray, nil
}

func (c *Client) ES_GetServicesStatus(from time.Time, to time.Time, elasticQuery ...elastic.Query) ([]*status.ServiceStatus, error) {
	var serviceStatusArray []*status.ServiceStatus

	t1, _ := time.Parse(timeLayout, "2019-01-07 16:08:00")
	s1 := &status.ServiceStatus{
		Id:              1,
		Interval:        30,
		Result:          true,
		Duration:        time.Second,
		ReqId:           "xxx",
		FailThreshold:   5,
		Message:         "ok",
		InsertTimestamp: t1,
	}

	t2, _ := time.Parse(timeLayout, "2019-01-07 16:09:00")
	s2 := &status.ServiceStatus{
		Id:              1,
		Interval:        60,
		Result:          true,
		Duration:        time.Second,
		ReqId:           "xxx",
		FailThreshold:   5,
		Message:         "ok",
		InsertTimestamp: t2,
	}

	t3, _ := time.Parse(timeLayout, "2019-01-07 12:04:00")
	s3 := &status.ServiceStatus{
		Id:              2,
		Interval:        30,
		Result:          true,
		Duration:        time.Second,
		ReqId:           "xxx",
		FailThreshold:   5,
		Message:         "ok",
		InsertTimestamp: t3,
	}

	t4, _ := time.Parse(timeLayout, "2019-01-07 12:04:30")
	s4 := &status.ServiceStatus{
		Id:              2,
		Interval:        30,
		Result:          false,
		Duration:        time.Second,
		ReqId:           "xxx",
		FailThreshold:   5,
		Message:         "ok",
		InsertTimestamp: t4,
	}

	t5, _ := time.Parse(timeLayout, "2019-01-07 12:05:00")
	s5 := &status.ServiceStatus{
		Id:              2,
		Interval:        30,
		Result:          true,
		Duration:        time.Second,
		ReqId:           "xxx",
		FailThreshold:   5,
		Message:         "ok",
		InsertTimestamp: t5,
	}

	t6, _ := time.Parse(timeLayout, "2019-01-07 13:04:30")
	s6 := &status.ServiceStatus{
		Id:              3,
		Interval:        30,
		Result:          true,
		Duration:        time.Second,
		ReqId:           "xxx",
		FailThreshold:   5,
		Message:         "ok",
		InsertTimestamp: t6,
	}

	t7, _ := time.Parse(timeLayout, "2019-01-07 13:05:00")
	s7 := &status.ServiceStatus{
		Id:              3,
		Interval:        30,
		Result:          true,
		Duration:        time.Second,
		ReqId:           "xxx",
		FailThreshold:   5,
		Message:         "ok",
		InsertTimestamp: t7,
	}

	serviceStatusArray = append(serviceStatusArray, s1, s2, s3, s4, s5, s6, s7)

	return serviceStatusArray, nil
}

func (c *Client) ES_SaveServiceStatus(s *status.ServiceStatus) error {
	// TODO
	fmt.Printf("ES_SaveServiceStatus - NOT IMPLEMENTED\n")
	return nil
}

func (c *Client) ES_DeleteServicesStatus(from time.Time, to time.Time) error {
	fmt.Printf("ES_DeleteServicesStatus - NOT IMPLEMENTED\n")
	return nil
}

func (c *Client) ES_GetAggregatedServiceStatusByID(from time.Time, to time.Time, serviceID int) (*status.AgregatedServiceStatus, error) {
	var serviceStatus *status.AgregatedServiceStatus

	if serviceID == 1 {
		from, _ := time.Parse(timeLayout, "2019-01-07 16:02:10")
		to, _ := time.Parse(timeLayout, "2019-01-07 16:05:35")
		serviceStatus = &status.AgregatedServiceStatus{
			Id:            "xaxWSCsxw",
			ServiceID:     1,
			Result:        true,
			Interval:      30,
			Aggregated:    10,
			AvgDuration:   time.Second,
			TimestampFrom: from,
			TimestampTo:   to,
		}
	} else if serviceID == 2 {
		from, _ := time.Parse(timeLayout, "2019-01-07 12:02:30")
		to, _ := time.Parse(timeLayout, "2019-01-07 12:03:00")
		serviceStatus = &status.AgregatedServiceStatus{
			Id:            "xaxsxAFWw",
			ServiceID:     2,
			Result:        true,
			Interval:      30,
			Aggregated:    1,
			AvgDuration:   time.Second,
			TimestampFrom: from,
			TimestampTo:   to,
		}
	} else if serviceID == 3 {
		serviceStatus = nil
	}

	return serviceStatus, nil
}

func (c *Client) ES_SaveAggregatedServiceStatus(s *status.AgregatedServiceStatus) error {
	// TODO
	fmt.Printf("ES_SaveAggregatedServiceStatus - NOT IMPLEMENTED,\n")
	return nil
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
			Type:   "email",
		}

		userNotifSettings = append(userNotifSettings, user1Notif)
	} else if checkId == 2 {
		// user1 email
		user1Notif := &notification.UserNotificationSettings{
			Target: "jardaID2@seznam.cz",
			Type:   "email",
		}
		// user3 email
		user2Notif := &notification.UserNotificationSettings{
			Target: "123456789ID2",
			Type:   "sms",
		}

		userNotifSettings = append(userNotifSettings, user1Notif)
		userNotifSettings = append(userNotifSettings, user2Notif)
	} else if checkId == 3 {
		// user1 email
		user1Notif := &notification.UserNotificationSettings{
			Target: "TomosID3@seznam.cz",
			Type:   "email",
		}

		userNotifSettings = append(userNotifSettings, user1Notif)

	} else if checkId == 4 {
		// user1 email
		user1Notif := &notification.UserNotificationSettings{
			Target: "456789854ID4",
			Type:   "sms",
		}

		userNotifSettings = append(userNotifSettings, user1Notif)
	}

	return userNotifSettings, nil
}

func (c *Client) SQL_GetServices(interval int) ([]*service.Service, error) {
	var services []*service.Service

	if interval == 30 {
		s1 := &service.Service{
			ID:            3,
			Type:          2,
			Interval:      30,
			FailThreshold: 5,
			Host:          "myhost",
			Metadata:      `{"id": 3,"target": "seznam.cz","port": 1234,"timeout": 5}`,
		}

		services = append(services, s1)
	}

	if interval == 60 {
		s1 := &service.Service{
			ID:            4,
			Type:          3,
			Interval:      60,
			FailThreshold: 5,
			Host:          "myhost",
			Metadata:      `{"id": 4,"target": "seznam.cz","timeout": 5}`,
		}

		s2 := &service.Service{
			ID:            5,
			Type:          1,
			Interval:      60,
			FailThreshold: 3,
			Host:          "myhost",
			Metadata: `{
	"id": 1,
	"port": 443,
	"target": "https://master.cz",
	"timeout": 5,
	"method": "GET",
	"query": "?var1=value1&var2=value2",
	"postData": [
		{
			"name": "var1",
			"value": "value1"
		}
	],
	"extraHeaders": [
		{
			"name": "MyHeader",
			"value": "My Value"
		}
	],
	"authEnabled": false,
	"authUsername": "admin",
	"authPassword": "adminPass",
	"contentCheckEnabled": false,
	"contentCheckString": "my_string",
	"allowedHttpStatusCodes": [
		200,
		201,
		403,
		404
	],
	"tlsSkipVerify": false,
	"tlsCheckCertificates": true,
	"tlsCertExpirationThreshold": 10
}`,
		}

		services = append(services, s1)
		services = append(services, s2)
	}

	return services, nil
}

func (c *Client) SQL_GetServiceDetails(checkID int) (*service.Service, error) {
	var serviceDetail *service.Service
	if checkID == 1 {
		serviceDetail = &service.Service{
			ID:            1,
			Host:          "myServer1",
			Target:        "web.myserver.com",
			Type:          2,
			FailThreshold: 5,
			Interval:      30,
		}
	} else if checkID == 2 {
		serviceDetail = &service.Service{
			ID:            2,
			Host:          "myWeb1",
			Target:        "webik.com",
			Type:          2,
			FailThreshold: 5,
			Interval:      30,
		}

	} else if checkID == 3 {
		serviceDetail = &service.Service{
			ID:            3,
			Host:          "bigServer",
			Target:        "seznam.com",
			Type:          2,
			FailThreshold: 3,
			Interval:      30,
		}

	} else if checkID == 4 {
		serviceDetail = &service.Service{
			ID:            4,
			Host:          "myICMPTestServer",
			Target:        "google.com",
			Type:          3,
			FailThreshold: 3,
			Interval:      30,
		}

	}

	return serviceDetail, nil
}

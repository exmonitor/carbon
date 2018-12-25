package service

import (
	"time"

	"github.com/pkg/errors"

	"github.com/exmonitor/firefly/database"
	"github.com/exmonitor/firefly/database/spec/status"
	"github.com/exmonitor/firefly/log"
	"github.com/exmonitor/firefly/notification"
)

type Config struct {
	DBClient      database.ClientInterface
	FetchInterval time.Duration

	TimeProfiling bool
	Logger        *log.Logger
}

type Service struct {
	dbClient      database.ClientInterface
	fetchInterval time.Duration

	timeProfiling bool

	logger *log.Logger
	// internals
	failedServiceDB map[int]FailedService // int is holder for check ID
	lastFetchTime   time.Time
}

// Create new Service
func New(conf Config) (*Service, error) {
	if conf.DBClient == nil {
		return nil, errors.Wrapf(invalidConfigError, "conf.DBClient must not be nil")
	}
	if conf.Logger == nil {
		return nil, errors.Wrapf(invalidConfigError, "conf.Logger must not be nil")
	}
	if conf.FetchInterval == 0 {
		return nil, errors.Wrapf(invalidConfigError, "conf.FetchInterval must not be zero")
	}

	newService := &Service{
		dbClient:      conf.DBClient,
		logger:        conf.Logger,
		fetchInterval: conf.FetchInterval,

		failedServiceDB: map[int]FailedService{},
		lastFetchTime:   time.Now().Add(-conf.FetchInterval),
	}

	return newService, nil
}

// make sure that the Loop is executed only once every x seconds defined in fecthInterval
func (s *Service) Boot() {

	// run tick goroutine
	tickChan := make(chan bool)
	s.logger.LogDebug("booting loop for interval %ds", int(s.fetchInterval.Seconds()))
	go intervalTick(int(s.fetchInterval.Seconds()), tickChan)

	// run infinite loop
	for {
		// wait until we reached another interval tick
		select {
		case <-tickChan:
			s.logger.Log("received tick, interval %ds", int(s.fetchInterval.Seconds()))
		}
		err := s.mainLoop()

		if err != nil {
			s.logger.LogError(err, "mainLoop failed")
		}
	}

}

func (s *Service) mainLoop() error {
	from := s.lastFetchTime
	to := time.Now()

	currentFailedServices, err := s.dbClient.ES_GetFailedServices(from, to, int(s.fetchInterval.Seconds()))
	if err != nil {
		return errors.Wrap(err, "failed to get currentFailedServices from DB")
	}
	s.logger.LogDebug("fetched %d failedServices for interval %ds", len(currentFailedServices), int(s.fetchInterval.Seconds()))

	// increase failCounter for existing ids or add id to the database
	for _, c := range currentFailedServices {
		// check if this id is already present in the list
		if failedService, ok := s.failedServiceDB[c.Id]; ok {
			// id is present increase failCheckDB, increase fail counter
			failedService.FailCounter += 1
			if failedService.FailCounter >= failedService.FailThreshold {
				// never count fails over threshold
				failedService.FailCounter = failedService.FailThreshold
				// check if we reached threshold and possible send notification
				s.maybeSendFailNotification(failedService)

				// if counter is over threshold we dont save as we dont need to increase the counter anymore
			} else {
				// safe back to localDB
				s.failedServiceDB[c.Id] = failedService
				s.logger.LogDebug("increasing failCounter for failedService ID:%d to %d", failedService.Id, failedService.FailCounter)
			}
		} else {
			s.failedServiceDB[c.Id] = FailedService{
				Id:            c.Id,
				FailCounter:   1,
				FailThreshold: c.FailThreshold,
				LastFailedMsg: c.Message,
				ResentEvery:   c.ResentEvery,
			}
			s.logger.LogDebug("adding new failedService with ID:%d to localDB", c.Id)
		}
	}

	// TODO, we dont instantly remove from failedDB if we got successful check but rather decrease counter,
	// TODO this can help avoid  hiding flapping alarms
	// reduce counter for missing checks
	for id, failedService := range s.failedServiceDB {
		if !status.Exists(currentFailedServices, id) {
			failedService.FailCounter -= 1
			// if failed counter drops to zero, than remove it from the failedCheckDb and send OK notification
			if failedService.FailCounter <= 0 {
				// remove check from db
				delete(s.failedServiceDB, id)
				// send OK notification
				s.sendOKNotification(failedService)
			} else {
				s.failedServiceDB[id] = failedService
				s.logger.LogDebug("decreasing fail counter for failedService ID:%d to %d", id, failedService.FailCounter)
			}
		}
	}

	return nil
}

// check if we need to send fail notification and do it
func (s *Service) maybeSendFailNotification(f FailedService) {
	// check if we already sent notification or not
	if !f.SentNotification {
		// we did not sent notification yet
		s.sendFailNotification(f)
	} else {
		// we sent notification Logalready
		// check if we should resent notification
		if f.SentNotificationTime.Add(f.ResentEvery).Before(time.Now()) {
			s.sendFailNotification(f)
		}
	}
}

// send FAIL notification
func (s *Service) sendFailNotification(f FailedService) {
	// log
	if !f.SentNotification {
		s.logger.LogDebug("sending FAIL notification for service ID %d", f.Id)
	} else {
		s.logger.LogDebug("resending FAIL notification for service ID %d after %s", f.Id, f.ResentEvery)
	}
	// set variables first
	f.SentNotification = true
	f.SentNotificationTime = time.Now()
	// save changes to map
	s.failedServiceDB[f.Id] = f

	// init notification settings
	notificationConfig := notification.Config{
		DBClient:  s.dbClient,
		ServiceID: f.Id,
		Failed:    true,
		Logger:    s.logger,
	}
	n, err := notification.New(notificationConfig)
	if err != nil {
		s.logger.LogError(err, "failed to create notification settings for service ID %d", f.Id)
	}
	// send notification in separate goroutine to avoid I/O block
	go n.Run()
}

// send OK notification
func (s *Service) sendOKNotification(f FailedService) {
	// log
	s.logger.LogDebug("sending OK notification for check %d", f.Id)

	// init notification settings
	notificationConfig := notification.Config{
		DBClient:  s.dbClient,
		ServiceID: f.Id,
		Failed:    false,
		Logger:    s.logger,
	}
	n, err := notification.New(notificationConfig)
	if err != nil {
		s.logger.LogError(err, "failed to create notification settings for service ID %d", f.Id)
	}
	// send notification in separate goroutine to avoid I/O block
	go n.Run()
}

// send true to tickChan every intervalSec
func intervalTick(intervalSec int, tickChan chan bool) {
	for {
		// extract amount of second and minutes from the now time
		_, min, sec := time.Now().Clock()
		// get sum of total secs in hour as intervals can be bigger than 59 sec
		totalSeconds := min*60 + sec

		// check if we hit the interval
		if totalSeconds%intervalSec == 0 {
			// send msg to the channel that we got tick
			tickChan <- true
			time.Sleep(time.Second)
		}
		// this is rough value, so we are testing 10 times per sec to not have big offset
		time.Sleep(time.Millisecond * 100)
	}
}

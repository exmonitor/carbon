package service

import (
	"github.com/giantswarm/project-lotus/carbon/status"
	"log"
	"time"
	"github.com/pkg/errors"
	"fmt"
)

const (
	fetchIntervalSec = 30
	fetchInterval = time.Second * fetchIntervalSec
)

type Config struct {
	DBClient CarbonDatabase
}

type Service struct {
	dbClient CarbonDatabase

	// internals
	failedChecksDb map[int]status.FailedCheck // int is holder for check ID
	lastFetchTime  time.Time
}

type CarbonDatabase interface {
	GetCheckResults(from time.Time, to time.Time) ([]*status.Status, error)
}

// Create new Service
func New(conf Config) (*Service, error) {
	if conf.DBClient == nil {
		return nil, invalidConfigError
	}

	newService := &Service{
		dbClient:       conf.DBClient,
		failedChecksDb: map[int]status.FailedCheck{},
		lastFetchTime:  time.Now().Add(-fetchInterval),
	}

	return newService, nil
}

// make sure that the Loop is executed only once every x seconds defined in fecthIntervalSec
func (s *Service) RunLoop() {

	// run tick goroutine
	tickChan := make(chan bool)
	go intervalTick(fetchIntervalSec, tickChan)

	// run infinite loop
	for {
		// wait until we reached another interval tick
		select {
		case <-tickChan:
			s.Log("received tick")
		}
		err := s.mainLoop()

		if err != nil {
			s.LogError("mainLoop failed", err)
		}
	}

}

func (s *Service) mainLoop() error {
	from := s.lastFetchTime
	to := time.Now()

	currentFailedChecks, err := s.dbClient.GetCheckResults(from, to)
	if err != nil {
		return errors.Wrap(err, "failed to get currentFailedChecks from db")
	}

	// increase failCounter for existing ids or add id to the database
	for _, c := range currentFailedChecks {
		// check if this id is already present in the list
		if failedCheck, ok := s.failedChecksDb[c.Id]; ok {
			// id is present increase failCheckDB, increase fail counter
			failedCheck.FailCounter +=1
			if failedCheck.FailCounter >= failedCheck.FailThreshold {
				// never count fails over threshold
				failedCheck.FailCounter = failedCheck.FailThreshold
				// check if we reached threshold and possible send notification
				s.maybeSendFailNotification(failedCheck)
			}


		} else {
			s.failedChecksDb[c.Id] = status.FailedCheck{
				Id:c.Id,
				FailCounter:1,
				FailThreshold:c.FailThreshold,
				LastFailedMsg:c.Message,
			}
		}
	}

	// TODO, we dont instantly remove from failedDB if we got sucessfull check but rather decrease counter,
	// TODO this can help avoid  hiding flapping alarms
	// reduce counter for missing checks
	for id, failedCheck := range s.failedChecksDb {
		if !status.StatusExists(currentFailedChecks, id) {
			failedCheck.FailCounter -= 1
			// if failed counter drops to zero, than remove it from the failedCheckDb and send OK notification
			if failedCheck.FailCounter <= 0 {
				delete(s.failedChecksDb, id)
			}
		}
	}



	return nil
}
// check if we need to send fail notification and do it
func (s *Service) maybeSendFailNotification(f status.FailedCheck) {
	// check if we already sent notification or not
	if !f.SentNotification {
		// we did not sent notification yet
		s.sendFailNotification(f)
	} else {
		// we sent notification already
		// check if we should resent notification
		if f.SentNotificationTime.Add(f.ResentEvery).Before(time.Now()) {
			s.sendFailNotification(f)
		}
	}
}
// send FAIL notification
func (s *Service) sendFailNotification(f status.FailedCheck)  {
	// log
	if !f.SentNotification {
		s.Log(fmt.Sprintf("sending FAIL notification for check %d",f.Id))
	} else {
		s.Log(fmt.Sprintf("sending FAIL notification for check %d after %s",f.Id, f.ResentEvery))
	}
	// set variables first
	f.SentNotification = true
	f.SentNotificationTime = time.Now()

	// TODO
}

// send OK notification
func (s *Service) sendOKNotification(f status.FailedCheck) {
	// log
	s.Log(fmt.Sprintf("sending OK notification for check %d",f.Id))

	// TODO
}


// returns true if its time to run the interval
func intervalTick(intervalSec int, tickChan chan bool) bool {
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
		//  this is rough value, so we are testing 10 times per sec to not have big offset
		time.Sleep(time.Millisecond * 100)
	}
}

func (s *Service) Log(msg string) {
	log.Printf("INFO|%s", msg)
}
func (s *Service) LogError(msg string, err error) {
	log.Printf("ERROR|%s|%s", msg, err)
}

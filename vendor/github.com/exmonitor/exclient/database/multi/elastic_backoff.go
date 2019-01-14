package multi

import (
	"context"
	"net/http"
	"time"

	"github.com/exmonitor/exlogger"
	"github.com/olivere/elastic"
)

const (
	elasticMaxRetry = 20
)

var elasticBackoffMin = time.Millisecond * 200
var elasticBackoffMax = time.Second * 60

type elasticBackoff struct {
	backoff elastic.Backoff
	logger  *exlogger.Logger
}

func NewElasticBackoff(logger *exlogger.Logger) *elasticBackoff {
	return &elasticBackoff{
		backoff: elastic.NewExponentialBackoff(elasticBackoffMin, elasticBackoffMax),
		logger:  logger,
	}
}


func (e *elasticBackoff) Retry(ctx context.Context, retry int, req *http.Request, resp *http.Response, err error) (time.Duration, bool, error) {

	// Stop after maxRetires
	if retry >= elasticMaxRetry {
		e.logger.LogError(executionFailedError, "elasticSearchRetrier failed  after %d retries", elasticMaxRetry)
		return 0, false, executionFailedError
	}

	e.logger.Log("retrying elasticSearch db request  %d/%d", retry, elasticMaxRetry)
	// Let the backoff strategy decide how long to wait and whether to stop
	wait, stop := e.backoff.Next(retry)
	return wait, stop, nil
}


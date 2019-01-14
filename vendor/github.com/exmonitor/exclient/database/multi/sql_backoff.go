package multi

import (
	"time"

	"github.com/cenkalti/backoff"

	"github.com/exmonitor/exlogger"
)

const (
	sqlMaxRetry = 15
)

var sqlBackoffMin = time.Millisecond * 500
var sqlBackoffMax = time.Second * 20
var sqlMaxElapsedTime = time.Minute * 120

type sqlBackoff struct {
	maxRetries uint64
	retryCount uint64
	underlying backoff.BackOff
	logger *exlogger.Logger
}

func NewSQLBackoff(logger *exlogger.Logger) backoff.BackOff {

	b := &backoff.ExponentialBackOff{
		InitialInterval:     sqlBackoffMin,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         sqlBackoffMax,
		MaxElapsedTime:      sqlMaxElapsedTime,
		Clock:               backoff.SystemClock,
	}

	b.Reset()

	s := &sqlBackoff{
		maxRetries: sqlMaxRetry,
		retryCount: 0,
		underlying: b,
		logger: logger,
	}

	return s
}

func (b *sqlBackoff) NextBackOff() time.Duration {
	if b.retryCount+1 >= b.maxRetries {
		return backoff.Stop
	}
	b.retryCount++

	b.logger.Log("retrying elasticSearch db request  %d/%d", b.retryCount, b.maxRetries)

	return b.underlying.NextBackOff()
}

func (b *sqlBackoff) Reset() {
	b.retryCount = 0
	b.underlying.Reset()
}

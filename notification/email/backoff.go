package email


import (
	"time"

	"github.com/cenkalti/backoff"
	"github.com/exmonitor/exlogger"
)

const (
	maxRetry = 15
)

var backoffMin = time.Millisecond * 500
var backoffMax = time.Second * 20
var maxElapsedTime = time.Minute * 120

type emailBackoff struct {
	maxRetries uint64
	retryCount uint64
	underlying backoff.BackOff
	logger *exlogger.Logger
}

func NewEmailBackoff(logger *exlogger.Logger) backoff.BackOff {

	b := &backoff.ExponentialBackOff{
		InitialInterval:     backoffMin,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         backoffMax,
		MaxElapsedTime:      maxElapsedTime,
		Clock:               backoff.SystemClock,
	}

	b.Reset()

	s := &emailBackoff{
		maxRetries: maxRetry,
		retryCount: 0,
		underlying: b,
		logger:     logger,
	}

	return s
}

func (b *emailBackoff) NextBackOff() time.Duration {
	if b.retryCount+1 >= b.maxRetries {
		return backoff.Stop
	}
	b.retryCount++

	b.logger.Log("retrying email request  %d/%d", b.retryCount, b.maxRetries)

	return b.underlying.NextBackOff()
}

func (b *emailBackoff) Reset() {
	b.retryCount = 0
	b.underlying.Reset()
}


package lock

import "time"

// NewExponentialBackoff returns new backoff with given values.
func NewExponentialBackoff(attempts int, initialInterval time.Duration, multiplier float64) Backoff {
	return &ExponentialBackoff{
		currentAttempt:  0,
		naxAttempts:     attempts,
		InitialInterval: initialInterval,
		Multiplier:      multiplier,
	}
}

// NewBackoff returns a backoff with default values.
// MaxAttempts: 3. InitialInterval: 1s, Nultiplier: 1.5
func NewBackoff() Backoff {
	naxAttempts := 3
	initialInterval := 1 * time.Second
	multiplier := 1.5
	return NewExponentialBackoff(naxAttempts, initialInterval, multiplier)
}

// NewEmptyBackoff returns a backoff with no attempts and no interval.
func newEmptyBackoff() Backoff {
	return &ExponentialBackoff{
		currentAttempt:  0,
		naxAttempts:     0,
		InitialInterval: time.Duration(0),
		Multiplier:      0,
	}
}

// Start reset current attempts to 0.
func (backoff *ExponentialBackoff) Start() {
	backoff.currentAttempt = 0
}

// Next calculates duration until next attempt should be run.
func (backoff *ExponentialBackoff) Next() *time.Duration {

	if backoff.currentAttempt >= backoff.naxAttempts {
		return nil
	}

	nextBackoff := backoff.InitialInterval
	cnt := 0
	for cnt < backoff.currentAttempt {
		nextBackoff = time.Duration(float64(nextBackoff) * backoff.Multiplier)
		cnt++
	}

	backoff.currentAttempt++
	return &nextBackoff
}

// Attempts returns number of current attempts.
func (backoff *ExponentialBackoff) Attempts() int {
	return backoff.currentAttempt
}

// MaxAttempts returns number of max attempts a backoff is confibured for.
func (backoff *ExponentialBackoff) MaxAttempts() int {
	return backoff.naxAttempts
}

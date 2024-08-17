package lock

import (
	"time"

	model "github.com/tommzn/utte-model"
)

// Lock is unused to obtain and release resource locks.
type Lock interface {

	// Obtain will try to get a lock for given resource by using a default backoff. s. NewBackoff() for details.
	Obtain(resourceId model.Identifier, clientId model.Identifier) (*Lease, error)

	// ObtainWithBackoff will try to lock passed resource and uses given backoff for retry.
	ObtainWithBackoff(resourceId model.Identifier, clientId model.Identifier, backoff Backoff) (*Lease, error)

	// ObtainWithNoWait will fail if a lock for given resource can not be created immedeately.
	ObtainWithNoWait(resourceId model.Identifier, clientId model.Identifier) (*Lease, error)

	// Release should be used to clean a lock if it's no longer required.
	// Locks expired after a defined time, so it's not mandatory to call this method - but this will keep a lock for a unnecessary long time and may block other clients.
	Release(lease *Lease) error
}

// Backoff is used for retry if a lock can not be obtained directly.
type Backoff interface {

	// Reset number of attempts to start from scratch.
	Start()

	// Next calculates durection a client should eait for next attempts to get a lock.
	Next() *time.Duration

	// Attempts returns the number of alreadz executed circles.
	Attempts() int

	// MaxAttempts returns the maximun number of attempts used by a backoff.
	MaxAttempts() int
}

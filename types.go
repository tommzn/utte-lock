package lock

import (
	"database/sql"
	"time"

	"github.com/tommzn/go-config"
	"github.com/tommzn/go-log"
	"github.com/tommzn/go-secrets"
	model "github.com/tommzn/utte-model"
)

// Lease is a temprary lock of a resource.
type Lease struct {

	// ResourceId is an identifier of a locked resource.
	ResourceId model.Identifier

	// ClientId is a source application id which obtains a lock for a resource.
	ClientId model.Identifier

	// Expiry, UTC timestamp when a lock gets invalid.
	Expiry time.Time

	// Sequence, serials number of a lockused to avoid resource changes by deprecated locks.
	Sequence int
}

// ExponentialBackoff is used to retry obtaining a lock.
type ExponentialBackoff struct {

	// CurrentAttempt is the number of attempts alreadz executed.
	currentAttempt int

	// MaxAttempts is the number of retries a backoff allows.
	naxAttempts int

	// InitialInterval is the initial wait duration for a second attempt.
	InitialInterval time.Duration

	// Multiplier is applied to wait interval at each attempt.
	Multiplier float64
}

// DatabaseLock is an instance of Lock using a database to manage lock entries.
type DatabaseLock struct {

	// Db is the current database connection.
	db *sql.DB

	// Retention defines retension time of a lock.
	retention time.Duration

	// conf hold current config.
	conf config.Config

	// SecretsManager is used to get credentials for database connection.
	secretsManager secrets.SecretsManager

	// logger to write log messages.
	logger log.Logger
}

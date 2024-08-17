package lock

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	model "github.com/tommzn/utte-model"
)

// Obtain will try to get a lock for given resource using a default backoff. See NewBackoff().
func (lock *DatabaseLock) Obtain(resourceId model.Identifier, clientId model.Identifier) (*Lease, error) {
	return lock.ObtainWithBackoff(resourceId, clientId, NewBackoff())
}

// ObtainWithNoWait will only run one attempt to get a lock.
func (lock *DatabaseLock) ObtainWithNoWait(resourceId model.Identifier, clientId model.Identifier) (*Lease, error) {
	return lock.ObtainWithBackoff(resourceId, clientId, newEmptyBackoff())
}

// ObtainWithBackoff will tru to get a lock and uses given backoff if first ot follow up attempts fail.
func (lock *DatabaseLock) ObtainWithBackoff(resourceId model.Identifier, clientId model.Identifier, backoff Backoff) (*Lease, error) {

	err := lock.connect()
	if err != nil {
		return nil, err
	}

	backoff.Start()
	for {

		expiry := lock.lockExpiry()

		err := lock.createLockEntry(resourceId, clientId, expiry)
		if err == nil {
			lock.logger.Debugf("Lock obtained for %s after %d attempts.", resourceId, backoff.Attempts())
			return lock.newLease(resourceId, clientId, expiry)
		}
		lock.logger.Info(err)

		nextBackOff := backoff.Next()
		if nextBackOff == nil {
			return nil, fmt.Errorf("Unable to get lock for resource %s after %d repetitions, reson: %s", resourceId, backoff.Attempts(), err)
		}
		time.Sleep((*nextBackOff))
	}
}

// Release deletes given lock and should be used after all operations for a resource has been performed.
// Sure, a lock will expire, but you should ensure efficient usage of resource locks.
func (lock *DatabaseLock) Release(lease *Lease) error {

	deleteStmt := `delete from "resource_locks" where "resource_id" = $1 and "client_id" = $2 and "sequence_no" = $3`
	_, err := lock.db.Exec(deleteStmt, lease.ResourceId, lease.ClientId, lease.Sequence)
	return err
}

// CreateLockEntry will try to get a lock. Either directly or by evaluating expiy time of a lock and may release an existing them to get a new one.
func (lock *DatabaseLock) createLockEntry(resourceId model.Identifier, clientId model.Identifier, expiry time.Time) error {

	if err := lock.insertLockEntry(resourceId, clientId, expiry, nil); err == nil {
		lock.logger.Debug("Initial lock insert succeeded for ", resourceId)
		return nil
	}

	if err := lock.assertExpiredLock(resourceId); err != nil {
		return err
	}

	tx, err := lock.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	if err := lock.cleanupExpiredLock(resourceId, tx); err != nil {
		lock.logger.Debugf("Attempt to delete expired lock failed for %s, reason: %s", resourceId, err)
		_ = tx.Rollback()
		return err
	}

	if err := lock.insertLockEntry(resourceId, clientId, expiry, tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	lock.logger.Debug("Expired lock cleaned up and new inserted for ", resourceId)
	return tx.Commit()
}

// InsertLockEntry will write a lock record to used database table. If there's alreadz an active transaction you can pass it.
func (lock *DatabaseLock) insertLockEntry(resourceId model.Identifier, clientId model.Identifier, expiry time.Time, tx *sql.Tx) error {

	insertStmt := `insert into "resource_locks"("resource_id", "client_id", "expiry") values($1, $2, $3)`
	var err error
	if tx != nil {
		_, err = tx.Exec(insertStmt, resourceId, clientId, expiry.Unix())
	} else {
		_, err = lock.db.Exec(insertStmt, resourceId, clientId, expiry.Unix())
	}
	if err != nil {
		return err
	}
	return nil
}

// AssertExpiredLock evaluates lock expiry for given resource. If a lock exists and is exoired it returbs without an error, otherwise it fails.
func (lock *DatabaseLock) assertExpiredLock(resourceId model.Identifier) error {

	now := currentTime().Unix()

	currentExpiry, err := lock.getLockExpiry(resourceId)
	if err != nil {
		lock.logger.Debugf("Unable to get expiry for %s, reason: %s", resourceId, err)
		return err
	}

	lock.logger.Debugf("Curren expiry %d, now %d", currentExpiry, now)
	if int64(currentExpiry) > now {
		errExp := fmt.Errorf("Lock not expiry for %s, %d", resourceId, currentExpiry)
		lock.logger.Debug(errExp)
		return errExp
	}

	return nil
}

// CleanupExpiredLock will remove an expired lock for given resource.
func (lock *DatabaseLock) cleanupExpiredLock(resourceId model.Identifier, tx *sql.Tx) error {

	res, err := tx.Exec(`delete from "resource_locks" WHERE "resource_id" = $1 and "expiry" < $2`, resourceId, currentTime().Unix())
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if res == nil {
		_ = tx.Rollback()
		return errors.New("No response for lock deletion received.")
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected != 1 {
		lock.logger.Debugf("Invalid number of effected rows (%d) for delete of %s", rowsAffected, resourceId)
		_ = tx.Rollback()
		return fmt.Errorf("Lock for %s alreadz exists and is not jet expired.", resourceId)
	}

	return nil
}

// GetLockSequence returns the sequence number of an existing resource lock.
func (lock *DatabaseLock) getLockSequence(resourceId model.Identifier) (int, error) {

	rows, err := lock.db.Query(`SELECT "sequence_no" FROM "resource_locks" WHERE "resource_id" = $1`, resourceId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var sequence int
	rows.Next()
	err = rows.Scan(&sequence)
	if err != nil {
		return 0, err
	}

	return sequence, nil
}

// GetLockExpiry returns expiry time for an existing resource lock.
func (lock *DatabaseLock) getLockExpiry(resourceId model.Identifier) (int, error) {

	rows, err := lock.db.Query(`SELECT "expiry" FROM "resource_locks" WHERE "resource_id" = $1`, resourceId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var expiry int
	rows.Next()
	err = rows.Scan(&expiry)
	if err != nil {
		return 0, err
	}

	return expiry, nil
}

// NewLease creates a new lease with given values.
func (lock *DatabaseLock) newLease(resourceId model.Identifier, clientId model.Identifier, expiry time.Time) (*Lease, error) {

	sequence, err := lock.getLockSequence(resourceId)
	if err != nil {
		return nil, err
	}

	return &Lease{
		ResourceId: resourceId,
		ClientId:   clientId,
		Expiry:     expiry,
		Sequence:   sequence,
	}, nil
}

// Connect will open a new database connebtion if none exists.
func (lock *DatabaseLock) connect() error {

	if lock.db != nil {
		return nil
	}

	db, err := postgresConnect(lock.conf, lock.secretsManager)
	if err != nil {
		return err
	}
	lock.db = db
	return nil
}

// LockExpiry calculates expiry time based on defined lock duration. Returns time in UTC.
func (lock *DatabaseLock) lockExpiry() time.Time {
	return currentTime().Add(lock.retention)
}

// CurrentTime returns current time in UTC.
func currentTime() time.Time {
	return time.Now().UTC()
}

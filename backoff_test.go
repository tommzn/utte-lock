package lock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type BackoffTestSuite struct {
	suite.Suite
}

func TestBackoffTestSuite(t *testing.T) {
	suite.Run(t, new(BackoffTestSuite))
}

func (suite *BackoffTestSuite) TestExponentialBackoff() {

	backoff := NewBackoff()
	suite.Equal(3, backoff.MaxAttempts())

	backoff.Start()

	wait01 := backoff.Next()
	suite.NotNil(wait01)
	suite.Equal(1*time.Second, *wait01)

	wait02 := backoff.Next()
	suite.NotNil(wait02)
	suite.Equal(1*time.Second+500*time.Millisecond, *wait02)

	wait03 := backoff.Next()
	suite.NotNil(wait03)
	suite.Equal(2*time.Second+250*time.Millisecond, *wait03)

	suite.Nil(backoff.Next())
}

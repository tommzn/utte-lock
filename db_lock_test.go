package lock

import (
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/suite"
)

type DbLockTestSuite struct {
	suite.Suite
	mig *migrate.Migrate
}

func TestDbLockTestSuite(t *testing.T) {
	suite.Run(t, new(DbLockTestSuite))
}

func (suite *DbLockTestSuite) SetupSuite() {

	conf := loadConfigForTest(nil)
	secretsManager := secretsManagerForTest()

	var err error
	suite.mig, err = DbMigrations(conf, secretsManager)
	suite.NotNil(suite.mig)
	suite.Nil(err)

	suite.True(MigrationSucceeded(suite.mig.Up()))
}

func (suite *DbLockTestSuite) TearDownSuite() {

	suite.NotNil(suite.mig)
	suite.mig.Down()
}

func (suite *DbLockTestSuite) TestLockWithNoWait() {

	resourceId := identifierForTest()
	clientId := identifierForTest()
	lock := lockForTest()

	lease1, err1 := lock.ObtainWithNoWait(resourceId, clientId)
	suite.Nil(err1)
	suite.NotNil(lease1)

	lease2, err2 := lock.ObtainWithNoWait(resourceId, clientId)
	suite.NotNil(err2)
	suite.Nil(lease2)

	time.Sleep(3 * time.Second)

	lease3, err3 := lock.ObtainWithNoWait(resourceId, clientId)
	suite.Nil(err3)
	suite.NotNil(lease3)
}

func (suite *DbLockTestSuite) TestLockWithRelease() {

	resourceId := identifierForTest()
	clientId := identifierForTest()
	lock := lockForTest()

	lease1, err1 := lock.ObtainWithNoWait(resourceId, clientId)
	suite.Nil(err1)
	suite.NotNil(lease1)

	suite.Nil(lock.Release(lease1))

	lease2, err2 := lock.ObtainWithNoWait(resourceId, clientId)
	suite.Nil(err2)
	suite.NotNil(lease2)

	suite.True(lease2.Sequence > lease1.Sequence)
}

func (suite *DbLockTestSuite) TestLockWitBackoff() {

	resourceId := identifierForTest()
	clientId := identifierForTest()
	lock := lockForTest()

	lease1, err1 := lock.ObtainWithNoWait(resourceId, clientId)
	suite.Nil(err1)
	suite.NotNil(lease1)

	lease2, err2 := lock.Obtain(resourceId, clientId)
	suite.Nil(err2)
	suite.NotNil(lease2)
}

func lockForTest() Lock {
	return &DatabaseLock{
		retention:      2 * time.Second,
		conf:           loadConfigForTest(nil),
		secretsManager: secretsManagerForTest(),
		logger:         loggerForTest(),
	}
}

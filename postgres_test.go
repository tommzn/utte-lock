package lock

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PostgresTestSuite struct {
	suite.Suite
}

func TestPostgresTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}

func (suite *PostgresTestSuite) TestConnect() {

	conf := loadConfigForTest(nil)
	secretsManager := secretsManagerForTest()

	db, err := postgresConnect(conf, secretsManager)
	suite.NotNil(db)
	suite.Nil(err)

}

package lock

import (
	config "github.com/tommzn/go-config"
	"github.com/tommzn/go-log"
	"github.com/tommzn/go-secrets"
	core "github.com/tommzn/utte-core"
	model "github.com/tommzn/utte-model"
)

func loadConfigForTest(fileName *string) config.Config {

	configFile := "fixtures/testconfig.yml"
	if fileName != nil {
		configFile = *fileName
	}
	configLoader := config.NewFileConfigSource(&configFile)
	config, _ := configLoader.Load()
	return config
}

func loggerForTest() log.Logger {
	return log.NewLogger(log.Debug, nil, nil)
}

func secretsManagerForTest() secrets.SecretsManager {

	secretsDict := make(map[string]string)
	secretsDict["POSTGRES_USER"] = "postgres"
	secretsDict["POSTGRES_PASSWORD"] = "postgres"
	return secrets.NewStaticSecretsManager(secretsDict)
}

func identifierForTest() model.Identifier {
	return core.NewIdentifier()
}

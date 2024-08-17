package lock

import (
	"database/sql"
	"fmt"

	"github.com/tommzn/go-config"
	"github.com/tommzn/go-secrets"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// PostgresConnect will tru to connect to a Postgres database. Settings have to be passed via config.
// Example config:
//
//	db:
//	  host: localhost
//	  port: 5432
//	  dbname: postgres
//
// Credentials are expected as POSTGRES_USER and POSTGRES_PASSWORD, passed via secrets manager.
func postgresConnect(conf config.Config, secretsManager secrets.SecretsManager) (*sql.DB, error) {

	host := conf.Get("db.host", nil)
	port := conf.Get("db.port", config.AsStringPtr("5432"))
	dbname := conf.Get("db.dbname", config.AsStringPtr("postgres"))

	user, err1 := secretsManager.Obtain("POSTGRES_USER")
	if err1 != nil {
		return nil, err1
	}
	password, err2 := secretsManager.Obtain("POSTGRES_PASSWORD")
	if err2 != nil {
		return nil, err2
	}

	conn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", *host, *port, *user, *password, *dbname)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	return db, db.Ping()
}

// DbMigrations creates a db migrate instance. Location of migration scripts have to be passed via config.
// Example:
//
//	db:
//	  migrations: "./db/migrations/"
func DbMigrations(conf config.Config, secretsManager secrets.SecretsManager) (*migrate.Migrate, error) {

	migrationsDir := conf.Get("db.migrations", config.AsStringPtr(("/db/migrations")))

	db, err := postgresConnect(conf, secretsManager)
	if err != nil {
		return nil, err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	return migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", *migrationsDir), "postgres", driver)
}

// MigrationSucceeded verifies than an error return from migration up is nil or at least no changes has happend.
func MigrationSucceeded(err error) bool {
	return err == nil || err == migrate.ErrNoChange
}

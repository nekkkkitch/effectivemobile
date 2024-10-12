package migrator

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type Migrator struct {
	srcDriver source.Driver
}

// Создаёт экземпляр Migrator с миграционными SQL файлами
func MustGetNewMigrator(sqlFiles embed.FS, dirName string) *Migrator {
	//создание драйвера источника с миграционными SQL файлами
	d, err := iofs.New(sqlFiles, dirName)
	if err != nil {
		fmt.Println(err)
	}
	return &Migrator{
		srcDriver: d,
	}
}

func (m *Migrator) ApplyMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("unable to create db instance: %v", err)
	}
	migrator, err := migrate.NewWithInstance("migration_embeded_sql_files", m.srcDriver, "psql_d", driver)
	if err != nil {
		return fmt.Errorf("unable to create migration: %v", err)
	}
	defer migrator.Close()
	if err = migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("unable to apply migrations %v", err)
	}
	return nil
}

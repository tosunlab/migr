package main

import (
	"github.com/tosunlab/migrasyon/internal/driver"
	"github.com/tosunlab/migrasyon/internal/driver/mysql"
	"github.com/tosunlab/migrasyon/internal/driver/postgres"
	"github.com/tosunlab/migrasyon/internal/driver/sqlite"
)

// drivers builds the registry this binary can reach, refusing a name registered twice.
func drivers() (*driver.Registry, error) {
	registry := driver.NewRegistry()

	if err := registry.Register(postgres.Name, postgres.New); err != nil {
		return nil, err
	}

	if err := registry.Register(sqlite.Name, sqlite.New); err != nil {
		return nil, err
	}

	if err := registry.Register(mysql.Name, mysql.New); err != nil {
		return nil, err
	}

	return registry, nil
}

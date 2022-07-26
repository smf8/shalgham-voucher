package database

import (
	"github.com/avast/retry-go/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

type DatabaseConfig struct {
	ConnectionAddress  string        `koanf:"connection-address"`
	RetryDelay         time.Duration `koanf:"note-expiry"`
	MaxRetry           uint          `koanf:"max-retry"`
	ConnectionLifetime time.Duration `koanf:"connection-lifetime"`
	MaxOpenConnections int           `koanf:"max-open-connections"`
	MaxIdleConnections int           `koanf:"max-idle-connections"`
}

// NewConnection will attempt connecting to database with given retry options.
func NewConnection(cfg DatabaseConfig) (db *gorm.DB, finalErr error) {
	finalErr = retry.Do(func() error {
		db, finalErr = gorm.Open(postgres.Open(cfg.ConnectionAddress), &gorm.Config{})
		if finalErr != nil {
			return finalErr
		}

		return nil
	}, retry.Delay(cfg.RetryDelay), retry.Attempts(cfg.MaxRetry))

	if finalErr != nil {
		return nil, finalErr
	}

	database, err := db.DB()
	if err != nil {
		return nil, err
	}

	database.SetMaxOpenConns(cfg.MaxOpenConnections)
	database.SetConnMaxLifetime(cfg.ConnectionLifetime)
	database.SetMaxIdleConns(cfg.MaxIdleConnections)

	return db, nil
}

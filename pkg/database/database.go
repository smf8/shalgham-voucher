package database

import (
	"time"

	"github.com/avast/retry-go/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	ConnectionAddress  string        `koanf:"connection-address"`
	RetryDelay         time.Duration `koanf:"retry-delay"`
	MaxRetry           uint          `koanf:"max-retry"`
	ConnectionLifetime time.Duration `koanf:"connection-lifetime"`
	MaxOpenConnections int           `koanf:"max-open-connections"`
	MaxIdleConnections int           `koanf:"max-idle-connections"`
	//1: silent,  2: error, 3: warning, 4: info
	LogLevel int `koanf:"log_level"`
}

// NewConnection will attempt connecting to database with given retry options.
func NewConnection(cfg DatabaseConfig) (db *gorm.DB, finalErr error) {

	finalErr = retry.Do(func() error {
		db, finalErr = gorm.Open(postgres.Open(cfg.ConnectionAddress),
			&gorm.Config{Logger: logger.Default.LogMode(logger.LogLevel(cfg.LogLevel))})
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

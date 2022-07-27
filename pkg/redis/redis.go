package redis

import (
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisConfig struct {
	Addresses       []string      `koanf:"addresses"`
	MasterName      string        `koanf:"master-name"`
	Password        string        `koanf:"password"`
	PoolSize        int           `koanf:"pool-size"`
	MinIdleConns    int           `koanf:"min-idle-conns"`
	DialTimeout     time.Duration `koanf:"dial-timeout"`
	ReadTimeout     time.Duration `koanf:"read-timeout"`
	WriteTimeout    time.Duration `koanf:"write-timeout"`
	PoolTimeout     time.Duration `koanf:"pool-timeout"`
	IdleTimeout     time.Duration `koanf:"idle-timeout"`
	MaxRetries      int           `koanf:"max-retries"`
	MinRetryBackoff time.Duration `koanf:"min-retry-backoff"`
	MaxRetryBackoff time.Duration `koanf:"max-retry-backoff"`
}

func New(cfg RedisConfig) (redis.Cmdable, func() error) {
	if cfg.MasterName == "" {
		return newNormal(cfg)
	}

	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:      cfg.MasterName,
		SentinelAddrs:   cfg.Addresses,
		Password:        cfg.Password,
		PoolSize:        cfg.PoolSize,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolTimeout:     cfg.PoolTimeout,
		IdleTimeout:     cfg.IdleTimeout,
		MinIdleConns:    cfg.MinIdleConns,
		MaxRetries:      cfg.MaxRetries,
		MinRetryBackoff: cfg.MinRetryBackoff,
		MaxRetryBackoff: cfg.MaxRetryBackoff,
	})

	return client, client.Close
}

func newNormal(cfg RedisConfig) (redis.Cmdable, func() error) {
	client := redis.NewClient(&redis.Options{
		Addr:            cfg.Addresses[0],
		PoolSize:        cfg.PoolSize,
		Password:        cfg.Password,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolTimeout:     cfg.PoolTimeout,
		IdleTimeout:     cfg.IdleTimeout,
		MinIdleConns:    cfg.MinIdleConns,
		MaxRetries:      cfg.MaxRetries,
		MinRetryBackoff: cfg.MinRetryBackoff,
		MaxRetryBackoff: cfg.MaxRetryBackoff,
	})

	return client, client.Close
}

package redis

import (
	"github.com/go-redis/redis/v8"
	"voucher/config"
)

func New(cfg config.Redis) (redis.Cmdable, func() error) {
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:      cfg.MasterName,
		SentinelAddrs:   cfg.Addresses,
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

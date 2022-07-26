package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/smf8/shalgham-voucher/pkg/database"
	"github.com/smf8/shalgham-voucher/pkg/redis"
	"github.com/smf8/shalgham-voucher/pkg/router"

	"github.com/sirupsen/logrus"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
)

const _Prefix = "VOUCHER_"

type Config struct {
	LogLevel     string                  `koanf:"log_level"`
	Server       router.ServerConfig     `koanf:"server"`
	Database     database.DatabaseConfig `koanf:"database"`
	WalletClient WalletClient            `koanf:"wallet_client"`
	Redis        redis.RedisConfig       `koanf:"redis"`
	VoucherCache VoucherCache            `koanf:"voucher_cache"`
}

type WalletClient struct {
	Timeout time.Duration `koanf:"timeout"`
	Debug   bool          `koanf:"debug"`
	BaseURL string        `koanf:"base_url"`
}

type VoucherCache struct {
	CronPattern string `koanf:"cron_pattern"`
}

var def = Config{
	LogLevel: "debug",
	Server: router.ServerConfig{
		Port:         ":8000",
		Debug:        true,
		NameSpace:    "voucher",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	},
	Database: database.DatabaseConfig{
		ConnectionAddress:  "postgresql://root@127.0.0.1:26257/voucherdb",
		RetryDelay:         time.Second,
		MaxRetry:           20,
		ConnectionLifetime: 30 * time.Minute,
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
		LogLevel:           1,
	},
	Redis: redis.RedisConfig{
		Addresses:       []string{"localhost:26379"},
		MasterName:      "mymaster",
		PoolSize:        0,
		MinIdleConns:    20,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolTimeout:     4 * time.Second,
		IdleTimeout:     5 * time.Minute,
		MaxRetries:      5,
		MinRetryBackoff: time.Second,
		MaxRetryBackoff: 3 * time.Second,
	},
	VoucherCache: VoucherCache{
		CronPattern: "@every 15s",
	},
	WalletClient: WalletClient{
		Timeout: 5 * time.Second,
		Debug:   true,
		BaseURL: "http://127.0.0.1:8001",
	},
}

func New() Config {
	var instance Config

	k := koanf.New(".")

	if err := k.Load(structs.Provider(def, "koanf"), nil); err != nil {
		logrus.Fatalf("error loading default: %s", err)
	}

	if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		logrus.Errorf("error loading file: %s", err)
	}

	if err := k.Load(env.Provider(_Prefix, ".", func(s string) string {
		parsedEnv := strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, _Prefix)), "__", "-")

		fmt.Println(parsedEnv)

		return strings.ReplaceAll(parsedEnv, "_", ".")
	}), nil); err != nil {
		logrus.Errorf("error loading environment variables: %s", err)
	}

	if err := k.Unmarshal("", &instance); err != nil {
		logrus.Fatalf("error unmarshalling config: %s", err)
	}

	logrus.Infof("following configuration is loaded:\n%+v", instance)

	return instance
}

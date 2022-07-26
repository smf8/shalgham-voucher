package config

import (
	"fmt"
	"github.com/smf8/arvan-voucher/pkg/database"
	"github.com/smf8/arvan-voucher/pkg/redis"
	"github.com/smf8/arvan-voucher/pkg/router"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
)

const _Prefix = "VOUCHER_"

type Config struct {
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
	CronDuration string `koanf:"cron_duration"`
}

var def = Config{
	Server: router.ServerConfig{
		Port:      ":8000",
		Debug:     true,
		NameSpace: "voucher",
	},
	Database: database.DatabaseConfig{
		ConnectionAddress:  "postgresql://root@127.0.0.1:26257/defaultdb",
		RetryDelay:         time.Second,
		MaxRetry:           20,
		ConnectionLifetime: 30 * time.Minute,
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
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
}

func New() Config {
	var instance Config

	k := koanf.New(".")

	if err := k.Load(structs.Provider(def, "koanf"), nil); err != nil {
		logrus.Fatalf("error loading default: %s", err)
	}

	if err := k.Load(file.Provider("config.yml"), yaml.Parser()); err != nil {
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

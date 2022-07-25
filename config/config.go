package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
)

const _Prefix = "WALLET_"

type Config struct {
	Port         string       `koanf:"port"`
	Debug        bool         `koanf:"debug"`
	Database     Database     `koanf:"database"`
	WalletClient WalletClient `koanf:"wallet_client"`
	Redis        Redis        `koanf:"redis"`
	VoucherCache VoucherCache `koanf:"voucher_cache"`
}

type WalletClient struct {
	Timeout time.Duration `koanf:"timeout"`
	Debug   bool          `koanf:"debug"`
	BaseURL string        `koanf:"base_url"`
}

type VoucherCache struct {
	CronDuration string `koanf:"cron_duration"`
}

type Redis struct {
	Addresses       []string      `koanf:"address"`
	MasterName      string        `koanf:"master-name"`
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

type Database struct {
	ConnectionAddress  string        `koanf:"connection-address"`
	RetryDelay         time.Duration `koanf:"note-expiry"`
	MaxRetry           uint          `koanf:"max-retry"`
	ConnectionLifetime time.Duration `koanf:"connection-lifetime"`
	MaxOpenConnections int           `koanf:"max-open-connections"`
	MaxIdleConnections int           `koanf:"max-idle-connections"`
}

var def Config = Config{
	Port: ":8000",
	Database: Database{
		ConnectionAddress:  "postgresql://smf8:owKoCRiy0_9epYxC9wQ8rg@free-tier7.aws-eu-west-1.cockroachlabs.cloud:26257/defaultdb?sslmode=verify-full&options=--cluster%3Dsag-to-rayanesh-2755",
		RetryDelay:         time.Second,
		MaxRetry:           20,
		ConnectionLifetime: 30 * time.Minute,
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
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

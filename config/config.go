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
	Port     string   `koanf:"port"`
	Debug    bool     `koanf:"debug"`
	Database Database `koanf:"database"`
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
	Port: ":8001",
	Database: Database{
		ConnectionAddress: "postgresql://smf8:owKoCRiy0_9epYxC9wQ8rg@free-tier7.aws-eu-west-1.cockroachlabs.cloud:26257/defaultdb?sslmode=verify-full&options=--cluster%3Dsag-to-rayanesh-2755",
		RetryDelay:        time.Second,
		MaxRetry:          20,
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

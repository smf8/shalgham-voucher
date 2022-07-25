package log

import (
	"github.com/sirupsen/logrus"
	"time"
)

func SetupLogger(lvl string) {
	logLevel, err := logrus.ParseLevel(lvl)
	if err != nil {
		logLevel = logrus.ErrorLevel
	}

	logrus.SetLevel(logLevel)

	if logLevel == logrus.DebugLevel {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	}
}

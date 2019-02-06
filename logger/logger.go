package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

var dataDogFormatter = &logrus.JSONFormatter{
	FieldMap: logrus.FieldMap{
		logrus.FieldKeyTime:  "timestamp",
		logrus.FieldKeyLevel: "level",
		logrus.FieldKeyMsg:   "message",
	},
}

func init() {
	var formatter logrus.Formatter = new(logrus.TextFormatter)
	if os.Getenv("ENV") != "" {
		formatter = dataDogFormatter
	}

	logger = logrus.New()
	logger.SetFormatter(formatter)
}

func Get() *logrus.Logger {
	return logger
}

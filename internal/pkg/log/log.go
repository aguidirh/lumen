package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

func New(level string) *logrus.Logger {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	l.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	l.SetLevel(logLevel)
	return l
}

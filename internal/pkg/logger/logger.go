package HosterLogger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type Log struct {
	*logrus.Logger
	ConfigSet bool
	Term      bool
}

func New() *Log {
	// Initialize Logrus' Struct to avoid de-ref pointer issues
	l := &Log{
		Logger: logrus.New(),
	}

	// Log as JSON instead of the default ASCII/text formatter.
	// l.Logger.SetFormatter(&logrus.JSONFormatter{})

	// Set the logger's output to /dev/null by default,
	// if no file was configured using the SetFileLocation()
	nullFile, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0666)
	if err != nil {
		// Handle the error if unable to open /dev/null
		panic(err)
	}
	l.Logger.SetOutput(nullFile)

	// Set log level
	logLevel := os.Getenv("HOSTER_LOG_LEVEL")
	if logLevel == "DEBUG" {
		l.Logger.SetLevel(logrus.DebugLevel)
	} else {
		l.Logger.SetLevel(logrus.InfoLevel)
	}

	// Report caller func
	// l.Logger.SetReportCaller(true)

	term := os.Getenv("TERM")
	if strings.HasPrefix(term, "xterm") || strings.HasPrefix(term, "term") {
		l.Term = true
	}

	return l
}

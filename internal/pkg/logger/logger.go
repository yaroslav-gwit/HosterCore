package HosterLogger

import (
	"HosterCore/internal/pkg/emojlog"
	FreeBSDLogger "HosterCore/internal/pkg/freebsd/logger"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type Log struct {
	*logrus.Logger
	ConfigSet bool
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
	l.Logger.SetLevel(logrus.DebugLevel)
	// Report caller func
	l.Logger.SetReportCaller(true)

	return l
}

func (l *Log) SetFileLocation(logLocation string) {
	file, err := os.OpenFile(logLocation, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		FreeBSDLogger.LoggerToSyslog(FreeBSDLogger.LOGGER_SRV_SCHEDULER, FreeBSDLogger.LOGGER_LEVEL_ERROR, "could not use this file for logging "+logLocation+", falling back to STDOUT")
	} else {
		l.Logger.SetOutput(file)
	}
}

// Test Info Func
func (l *Log) Info(value interface{}) {
	term := os.Getenv("TERM")
	if strings.HasPrefix(term, "xterm") || strings.HasPrefix(term, "term") {
		stringValue := fmt.Sprintf("%s", value)
		emojlog.PrintLogLine(emojlog.Info, stringValue)
	}

	l.Logger.Info(value)
}

// Test Error Func
func (l *Log) Error(value interface{}) {
	term := os.Getenv("TERM")
	if strings.HasPrefix(term, "xterm") || strings.HasPrefix(term, "term") {
		stringValue := fmt.Sprintf("%s", value)
		emojlog.PrintLogLine(emojlog.Error, stringValue)
	}

	l.Logger.Error(value)
}

// Test Info Func
func (l *Log) InfoToFile(value interface{}) {
	l.Logger.Info(value)
}

// Test Error Func
func (l *Log) ErrorToFile(value interface{}) {
	l.Logger.Error(value)
}

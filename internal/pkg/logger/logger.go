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
	Logger       *logrus.Logger
	FileLocation string
}

// Log as JSON instead of the default ASCII/text formatter.
// log.SetFormatter(&logrus.JSONFormatter{})
func New() *Log {
	// Initialize Logrus' Struct
	l := &Log{
		Logger: logrus.New(),
	}

	// Set log output
	l.Logger.SetOutput(os.Stdout)
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
		l.FileLocation = logLocation
	}
}

// Test Info Func
func (l *Log) Info(value interface{}) {
	term := os.Getenv("TERM")
	if strings.HasPrefix(term, "xterm") || strings.HasPrefix(term, "term") {
		stringValue := fmt.Sprintf("%s", value)
		emojlog.PrintLogLine(emojlog.Info, stringValue)
	}

	if len(l.FileLocation) > 1 {
		l.Logger.Info(value)
	}
}

// Test Error Func
func (l *Log) Error(value interface{}) {
	term := os.Getenv("TERM")
	if strings.HasPrefix(term, "xterm") || strings.HasPrefix(term, "term") {
		stringValue := fmt.Sprintf("%s", value)
		emojlog.PrintLogLine(emojlog.Error, stringValue)
	}

	if len(l.FileLocation) > 1 {
		l.Logger.Info(value)
	}
}

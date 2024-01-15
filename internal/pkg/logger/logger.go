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
	Out          *logrus.Logger
	Err          *logrus.Logger
	File         *logrus.Logger
	FileLocation string
}

var logger *Log

func init() {
	// Log as JSON instead of the default ASCII/text formatter.
	// log.SetFormatter(&logrus.JSONFormatter{})
	// logger.SetDefaultValues()
}
func New() *Log {
	logger.SetDefaultValues()
	return logger
}

func (l *Log) SetDefaultValues() *Log {
	// Set log outputs
	l.Err.SetOutput(os.Stderr)
	l.Out.SetOutput(os.Stdout)
	l.File.SetOutput(os.Stdout)

	// Set log level
	l.Err.SetLevel(logrus.DebugLevel)
	l.Out.SetLevel(logrus.DebugLevel)
	l.File.SetLevel(logrus.DebugLevel)

	// Report caller func
	l.Err.SetReportCaller(true)
	l.Out.SetReportCaller(true)
	l.File.SetReportCaller(true)

	return l
}

func (l *Log) SetFileLocation(logLocation string) {
	file, err := os.OpenFile(logLocation, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		FreeBSDLogger.LoggerToSyslog(FreeBSDLogger.LOGGER_SRV_SCHEDULER, FreeBSDLogger.LOGGER_LEVEL_ERROR, "could not use this file for logging "+logLocation+", falling back to STDOUT")
	} else {
		l.File.SetOutput(file)
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

	if len(logger.FileLocation) > 1 {
		l.File.Info(value)
	}
}

// Test Error Func
func (l *Log) Error(value interface{}) {
	term := os.Getenv("TERM")
	if strings.HasPrefix(term, "xterm") || strings.HasPrefix(term, "term") {
		stringValue := fmt.Sprintf("%s", value)
		emojlog.PrintLogLine(emojlog.Error, stringValue)
	}

	if len(logger.FileLocation) > 1 {
		l.File.Info(value)
	}
}

package HosterLogger

import (
	"HosterCore/internal/pkg/emojlog"
	FreeBSDLogger "HosterCore/internal/pkg/freebsd/logger"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var logOut = logrus.New()
var logErr = logrus.New()
var logFile = logrus.New()
var logFileSet = false

func init() {
	// Log as JSON instead of the default ASCII/text formatter.
	// log.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	logOut.SetOutput(os.Stdout)
	logErr.SetOutput(os.Stderr)
	logFile.SetOutput(os.Stdout)

	// Set log level
	logOut.SetLevel(logrus.DebugLevel)
	logErr.SetLevel(logrus.DebugLevel)
	logFile.SetLevel(logrus.DebugLevel)

	// Report caller func
	logOut.SetReportCaller(true)
	logErr.SetReportCaller(true)
	logFile.SetReportCaller(true)
}

func SetFileLocation(logLocation string) {
	file, err := os.OpenFile(logLocation, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		FreeBSDLogger.LoggerToSyslog(FreeBSDLogger.LOGGER_SRV_SCHEDULER, FreeBSDLogger.LOGGER_LEVEL_ERROR, "could not use this file for logging "+logLocation+", falling back to STDOUT")
	} else {
		logFile.SetOutput(file)
		logFileSet = true
	}
}

func Info(value interface{}) {
	term := os.Getenv("TERM")
	if strings.HasPrefix(term, "xterm") || strings.HasPrefix(term, "term") {
		stringValue := fmt.Sprintf("%s", value)
		emojlog.PrintLogLine(emojlog.Info, stringValue)
	}

	if logFileSet {
		logFile.Info(value)
	}
}

func Error(value interface{}) {
	term := os.Getenv("TERM")
	if strings.HasPrefix(term, "xterm") || strings.HasPrefix(term, "term") {
		stringValue := fmt.Sprintf("%s", value)
		emojlog.PrintLogLine(emojlog.Error, stringValue)
	}

	if logFileSet {
		logFile.Error(value)
	}
}

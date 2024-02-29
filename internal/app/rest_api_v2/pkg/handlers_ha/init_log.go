package HandlersHA

// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

import (
	FreeBSDLogger "HosterCore/internal/pkg/freebsd/logger"
	"os"

	"github.com/sirupsen/logrus"
)

var intLog = logrus.New() // internal HA log to hoster_ha.log, the variable name was set to this in order to avoid any conflicts with the global RestAPI logger

func init() {
	logFile := HA_LOG_LOCATION

	// Log as JSON instead of the default ASCII/text formatter.
	// log.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	intLog.SetOutput(os.Stdout)

	// Log to file, but fallback to STDOUT if something goes wrong
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		FreeBSDLogger.LoggerToSyslog(FreeBSDLogger.LOGGER_SRV_SCHEDULER, FreeBSDLogger.LOGGER_LEVEL_ERROR, "DNS_SERVER: could not use this file for logging "+logFile+", falling back to STDOUT")
	} else {
		intLog.SetOutput(file)
	}

	intLog.SetLevel(logrus.DebugLevel)
	// log.SetReportCaller(true)
}

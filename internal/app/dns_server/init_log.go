// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package main

import (
	FreeBSDLogger "HosterCore/internal/pkg/freebsd/logger"
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	// Ignore logging if version was requested
	args := os.Args
	if len(args) > 1 {
		res := os.Args[1]
		if res == "version" || res == "v" || res == "--version" || res == "-v" {
			return
		}
	}

	logFile := os.Getenv("LOG_FILE")

	// Log as JSON instead of the default ASCII/text formatter.
	// log.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)

	// Log to file, but fallback to STDOUT if something goes wrong
	if len(logFile) > 2 {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			FreeBSDLogger.LoggerToSyslog(FreeBSDLogger.LOGGER_SRV_SCHEDULER, FreeBSDLogger.LOGGER_LEVEL_ERROR, "DNS_SERVER: could not use this file for logging "+logFile+", falling back to STDOUT")
		} else {
			log.SetOutput(file)
		}
	}

	log.SetLevel(logrus.DebugLevel)
	// log.SetReportCaller(true)
}

// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package main

import (
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	FreeBSDLogger "HosterCore/internal/pkg/freebsd/logger"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	// Do not set-up logging, if someone is calling in to get a binary version
	args := os.Args
	if len(args) > 1 {
		res := os.Args[1]
		if res == "version" || res == "v" || res == "--version" || res == "-v" {
			return
		}
	}

	// If log folder does not exist, create it
	if _, err := os.Stat(CarpUtils.LOG_FOLDER); os.IsNotExist(err) {
		err := os.MkdirAll(CarpUtils.LOG_FOLDER, 0755)
		if err != nil {
			FreeBSDLogger.LoggerToSyslog(FreeBSDLogger.LOGGER_SRV_SCHEDULER, FreeBSDLogger.LOGGER_LEVEL_ERROR, "HA_CARP: could not create log folder "+CarpUtils.LOG_FOLDER)
		}
	}

	logFile := os.Getenv("LOG_FILE") // Logs usually reside here (if not overridden): /opt/hoster-core/logs/ha_carp.log
	if len(logFile) < 2 {
		logFile = CarpUtils.LOG_FILE
	}

	// Log as JSON instead of the default ASCII/text formatter.
	// log.SetFormatter(&logrus.JSONFormatter{})

	// Rotate the log file if it gets bigger than 10MB
	err := rotateLogs(logFile, 10*1024*1024)
	if err != nil {
		FreeBSDLogger.LoggerToSyslog(FreeBSDLogger.LOGGER_SRV_SCHEDULER, FreeBSDLogger.LOGGER_LEVEL_ERROR, "HA_CARP: could not rotate log file "+logFile)
	}

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)

	// Log to file, but fallback to STDOUT if something goes wrong
	if len(logFile) > 2 {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			FreeBSDLogger.LoggerToSyslog(FreeBSDLogger.LOGGER_SRV_SCHEDULER, FreeBSDLogger.LOGGER_LEVEL_ERROR, "HA_CARP: could not use this file for logging "+logFile+", falling back to STDOUT")
		} else {
			log.SetOutput(file)
		}
	}

	log.SetLevel(logrus.DebugLevel)
	// log.SetReportCaller(true)
}

func rotateLogs(filePath string, maxSize int64) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No file exists yet, so no need to rotate
		}
		return err
	}

	// If the file is bigger than maxSize, rotate it
	if fileInfo.Size() >= maxSize {
		timestamp := time.Now().Format("20060102-150405")
		newFileName := fmt.Sprintf("%s.%s.log", filePath, timestamp)
		err := os.Rename(filePath, newFileName)
		if err != nil {
			return err
		}

		// Create a new log file
		_, err = os.Create(filePath)
		if err != nil {
			return err
		}
	}

	return nil
}

// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterLogger

import (
	FreeBSDLogger "HosterCore/internal/pkg/freebsd/logger"
	"os"
)

func (l *Log) SetFileLocation(logLocation string) {
	file, err := os.OpenFile(logLocation, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		FreeBSDLogger.LoggerToSyslog(FreeBSDLogger.LOGGER_SRV_SCHEDULER, FreeBSDLogger.LOGGER_LEVEL_ERROR, "could not use this file for logging "+logLocation+", falling back to STDOUT")
	} else {
		l.Logger.SetOutput(file)
	}
}

package handlers

import (
	MiddlewareLogging "HosterCore/internal/app/rest_api_v2/pkg/middleware/logging"
)

var log *MiddlewareLogging.Log

func SetLogConfig(l *MiddlewareLogging.Log) {
	log = l
}

func init() {
	// log = MiddlewareLogging.Configure(logrus.DebugLevel)
}

package HosterJail

import HosterLogger "HosterCore/internal/pkg/logger"

var log = HosterLogger.New()

// Override the logger for this package
func SetLogger(l *HosterLogger.Log) {
	log = l
}

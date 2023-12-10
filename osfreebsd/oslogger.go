package osfreebsd

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const (
	LOGGER_SRV_REST        = "       HOSTER_REST"
	LOGGER_SRV_HA_REST     = "    HOSTER_HA_REST"
	LOGGER_SRV_HA_WATCHDOG = "HOSTER_HA_WATCHDOG"
)

const (
	LOGGER_LEVEL_DEBUG   = "  DEBUG"
	LOGGER_LEVEL_ERROR   = "  ERROR"
	LOGGER_LEVEL_INFO    = "   INFO"
	LOGGER_LEVEL_CHANGE  = " CHANGE"
	LOGGER_LEVEL_SUCCESS = "SUCCESS"
	LOGGER_LEVEL_WARNING = "WARNING"
)

func LoggerToSyslog(service string, level string, message string) error {
	logMessage := fmt.Sprintf("%s: %s", level, message)
	out, err := exec.Command("logger", "-t", service, logMessage).CombinedOutput()
	if err != nil {
		errValue := strings.TrimSpace(string(out)) + "; " + err.Error()
		return errors.New(errValue)
	}

	return nil
}

func LoggerToFile(service string, level string, message string, fileLocation string) error {
	logMessage := fmt.Sprintf("%s: %s", level, message)
	out, err := exec.Command("logger", "-t", service, logMessage).CombinedOutput()
	if err != nil {
		errValue := strings.TrimSpace(string(out)) + "; " + err.Error()
		return errors.New(errValue)
	}

	return nil
}

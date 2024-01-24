package MiddlewareLogging

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var lock sync.RWMutex

type Log struct {
	*logrus.Logger
	RespError      bool
	ErrorMessage   string
	RespDebug      bool
	DebugMessage   string
	InfoMessage    string
	HttpStatusCode int
}

func Configure(level logrus.Level) *Log {
	lock.Lock()
	defer lock.Unlock()

	l := &Log{
		Logger:    logrus.New(),
		RespError: false,
	}

	logStdOut := os.Getenv("LOG_STDOUT")
	logFile := os.Getenv("LOG_FILE")

	// Log as JSON instead of the default ASCII/text formatter.
	l.SetFormatter(&logrus.JSONFormatter{})
	// Output to stdout instead of the default stderr
	l.SetOutput(os.Stdout)

	// Log to file, but fallback to STDOUT if something goes wrong
	if logStdOut == "false" && len(logFile) > 2 {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			_ = 0
		} else {
			l.SetOutput(file)
		}
	}

	l.SetLevel(level)
	l.SetReportCaller(false)

	return l
}

func (log *Log) revertToNormalLevel() {
	lock.Lock()
	defer lock.Unlock()

	log.RespError = false
	log.RespDebug = false
}

func (log *Log) SetErrorTrue() {
	lock.Lock()
	defer lock.Unlock()

	log.RespError = true
}

func (log *Log) SetErrorMessage(message string) {
	lock.Lock()
	defer lock.Unlock()

	log.RespError = true
	log.ErrorMessage = message
}

func (log *Log) SetDebugTrue() {
	lock.Lock()
	defer lock.Unlock()

	log.RespDebug = true
}

func (log *Log) SetDebugMessage(message string) {
	lock.Lock()
	defer lock.Unlock()

	log.RespDebug = true
	log.DebugMessage = message
}

func (log *Log) LogResponses(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()
		next.ServeHTTP(w, r)

		if log.RespDebug {
			log.WithFields(logrus.Fields{
				"method":         r.Method,
				"status_code":    log.HttpStatusCode,
				"url":            r.URL.Path,
				"client_address": r.RemoteAddr,
				"latency":        fmt.Sprintf("%dms", time.Since(timeStart).Milliseconds()),
			}).Debug(log.DebugMessage)
			log.revertToNormalLevel()

			return
		}

		if log.RespError {
			log.WithFields(logrus.Fields{
				"method":         r.Method,
				"status_code":    log.HttpStatusCode,
				"url":            r.URL.Path,
				"client_address": r.RemoteAddr,
				"latency":        fmt.Sprintf("%dms", time.Since(timeStart).Milliseconds()),
			}).Error(log.ErrorMessage)
			log.revertToNormalLevel()

			return
		}

		infoMessage := "success"
		if len(log.InfoMessage) > 0 {
			infoMessage = log.InfoMessage
		}
		log.WithFields(logrus.Fields{
			"method":         r.Method,
			"status_code":    log.HttpStatusCode,
			"url":            r.URL.Path,
			"client_address": r.RemoteAddr,
			"latency":        fmt.Sprintf("%dms", time.Since(timeStart).Milliseconds()),
		}).Info(infoMessage)
	})
}

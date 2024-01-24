package main

import (
	MiddlewareLogging "HosterCore/internal/app/rest_api_v2/middleware/logging"
	FreeBSDLogger "HosterCore/internal/pkg/freebsd/logger"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	logStdOut := os.Getenv("LOG_STDOUT")
	logFile := os.Getenv("LOG_FILE")

	// Log as JSON instead of the default ASCII/text formatter.
	// log.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)

	// Log to file, but fallback to STDOUT if something goes wrong
	if logStdOut == "false" && len(logFile) > 2 {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			FreeBSDLogger.LoggerToSyslog(FreeBSDLogger.LOGGER_SRV_SCHEDULER, FreeBSDLogger.LOGGER_LEVEL_ERROR, "DNS_SERVER: could not use this file for logging "+logFile+", falling back to STDOUT")
		} else {
			log.SetOutput(file)
		}
	}

	log.SetLevel(logrus.DebugLevel)
	log.SetReportCaller(true)
}

var logApi *MiddlewareLogging.Log

func main() {
	logApi = MiddlewareLogging.Configure(logrus.DebugLevel)

	r := mux.NewRouter()
	r.HandleFunc("/api/v2/", home)
	r.HandleFunc("/api/v2", home)
	r.HandleFunc("/api/v2/health", home)
	r.HandleFunc("/api/v2/health/", home)

	http.Handle("/", r)
	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:4000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("could not start the REST API server: " + err.Error())
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Category: %v\n", vars["category"])
}

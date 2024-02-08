package main

import (
	"HosterCore/internal/app/rest_api_v2/pkg/handlers"
	MiddlewareLogging "HosterCore/internal/app/rest_api_v2/pkg/middleware/logging"
	FreeBSDLogger "HosterCore/internal/pkg/freebsd/logger"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	_ "github.com/swaggo/http-swagger/example/gorilla/docs" // docs is generated by Swag CLI, you have to import it.
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

var logInternal = logrus.New()
var log *MiddlewareLogging.Log

func init() {
	logStdOut := os.Getenv("LOG_STDOUT")
	logFile := os.Getenv("LOG_FILE")

	// Log as JSON instead of the default ASCII/text formatter.
	// log.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	logInternal.SetOutput(os.Stdout)

	// Log to file, but fallback to STDOUT if something goes wrong
	if logStdOut == "false" && len(logFile) > 2 {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			FreeBSDLogger.LoggerToSyslog(FreeBSDLogger.LOGGER_SRV_SCHEDULER, FreeBSDLogger.LOGGER_LEVEL_ERROR, "DNS_SERVER: could not use this file for logging "+logFile+", falling back to STDOUT")
		} else {
			logInternal.SetOutput(file)
		}
	}

	logInternal.SetLevel(logrus.DebugLevel)
	// logInternal.SetReportCaller(true)
}

// var log *MiddlewareLogging.Log

// @title Hoster Node REST API Docs
// @version 2.0
// @description REST API documentation for the `Hoster` nodes. This HTTP endpoint is located directly on the `Hoster` node.<br>Please, take some extra care with the things you execute here, because many of them can be destructive and non-revertible (e.g. vm destroy, snapshot rollback, host reboot, etc).
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /api/v2
func main() {
	// log = MiddlewareLogging.Configure(logrus.DebugLevel)
	r := mux.NewRouter()

	// Middleware -> Logging
	log = MiddlewareLogging.Configure(logrus.DebugLevel)
	handlers.SetLogConfig(log)
	r.Use(log.LogResponses)

	// Health checks
	r.HandleFunc("/api/v2/health", handlers.HealthCheck).Methods("GET")
	r.HandleFunc("/api/v2/health/auth/ha", handlers.HealthCheckHaAuth).Methods("GET")
	r.HandleFunc("/api/v2/health/auth/any", handlers.HealthCheckAnyAuth).Methods("GET")
	r.HandleFunc("/api/v2/health/auth/regular", handlers.HealthCheckRegularAuth).Methods("GET")
	// Host
	r.HandleFunc("/api/v2/host/info", handlers.HostInfo).Methods(http.MethodGet)
	// Jails
	r.HandleFunc("/api/v2/jail/all", handlers.JailList).Methods(http.MethodGet)
	r.HandleFunc("/api/v2/jail/templates", handlers.JailListTemplates).Methods(http.MethodGet)
	r.HandleFunc("/api/v2/jail/info/{jail_name}", handlers.JailInfo).Methods(http.MethodGet)
	r.HandleFunc("/api/v2/jail/start/{jail_name}", handlers.JailStart).Methods(http.MethodPost)
	r.HandleFunc("/api/v2/jail/stop/{jail_name}", handlers.JailStop).Methods(http.MethodPost)
	r.HandleFunc("/api/v2/jail/destroy/{jail_name}", handlers.JailDestroy).Methods(http.MethodDelete)
	r.HandleFunc("/api/v2/jail/deploy", handlers.JailDeploy).Methods(http.MethodPost)

	// Swagger docs
	r.PathPrefix("/api/v2/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/api/v2/swagger.json"), //The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods("GET")
	// Define a route to serve the static file
	r.HandleFunc("/api/v2/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		log.SetStatusCode(http.StatusOK)
		http.ServeFile(w, r, "docs/swagger.json")
	})
	// Catch-all route for 404 errors
	r.NotFoundHandler = r.NewRoute().HandlerFunc(handlers.NotFoundHandler).GetHandler()

	logInternal.Info("The REST APIv2 is bound to :4000")
	http.Handle("/", r)
	srv := &http.Server{
		Addr:         "0.0.0.0:4000",
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	err := srv.ListenAndServe()
	if err != nil {
		logInternal.Fatal("could not start the REST API server: " + err.Error())
	}
}

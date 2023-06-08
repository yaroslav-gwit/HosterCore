package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprint(w, "hoster metrics exporter\ngo to /metrics to see the exported metrics")
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metricsText := concatMetrics()
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprint(w, metricsText)
		clearMetricsList()
	})

	http.ListenAndServe(":8080", nil)
}

var endsWithNewline = regexp.MustCompile(".*\n$")

func generateMetrics() string {
	metricsText := "# HELP hello_world_counter Number of times 'Hello, World!' has been requested.\n"
	metricsText = metricsText + "# TYPE hello_world_counter counter\n"
	metricsText = metricsText + "hello_world_counter 1\n"

	if !endsWithNewline.MatchString(metricsText) {
		metricsText = metricsText + "\n"
	}

	return metricsText
}

func generateMetrics1() string {
	metricsText := "# HELP hello_world_counter Number of times 'Hello, World!' has been requested.\n"
	metricsText = metricsText + "# TYPE hello_world_counter counter\n"
	metricsText = metricsText + "hello_world_counter 2"

	if !endsWithNewline.MatchString(metricsText) {
		metricsText = metricsText + "\n"
	}

	return metricsText
}

func generateMetrics2() string {
	metricsText := "# HELP hello_world_counter Number of times 'Hello, World!' has been requested.\n"
	metricsText = metricsText + "# TYPE hello_world_counter counter\n"
	metricsText = metricsText + "hello_world_counter 3\n"

	if !endsWithNewline.MatchString(metricsText) {
		metricsText = metricsText + "\n"
	}

	return metricsText
}

func getNodeExporterMetrics() string {
	url := "http://192.168.100.8:9100/metrics"

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to make the HTTP GET request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read the response body: %v", err)
	}

	metricsString := string(body)
	return metricsString
}

func concatMetrics() string {
	var metricsTextFinal string
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		metricsText := generateMetrics()
		addMetricsToList(metricsText)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		metricsText := generateMetrics1()
		addMetricsToList(metricsText)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		metricsText := generateMetrics2()
		addMetricsToList(metricsText)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		metricsText := getNodeExporterMetrics()
		addMetricsToList(metricsText)
		wg.Done()
	}()

	wg.Wait()

	for i, v := range metricsList {
		if i == 0 {
			metricsTextFinal = v
		} else {
			metricsTextFinal += v
		}
	}

	return metricsTextFinal
}

var metricsList []string
var metricsListMutex sync.Mutex

func addMetricsToList(metricsText string) {
	metricsListMutex.Lock()
	metricsList = append(metricsList, metricsText)
	metricsListMutex.Unlock()
}

func clearMetricsList() {
	metricsListMutex.Lock()
	metricsList = []string{}
	metricsListMutex.Unlock()
}

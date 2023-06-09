package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
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

type DiskInfo struct {
	queueLength     int
	opsPerSec       int
	readsPerSecOp   int
	readsPerSecKb   int
	readsTimePerOp  float32
	writesPerSecOp  int
	writesPerSecKb  int
	writesTimePerOp float32
	busyPercent     float32
	deviceName      string
}

func getGstatMetrics() string {
	cmd := exec.Command("gstat", "-bp", "-I 850000")

	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to execute gstat: %v", err)
	}

	diskInfo := []DiskInfo{}
	reSplitSpace := regexp.MustCompile(`\s+`)
	linesList := []string{}

	for i, v := range strings.Split(string(output), "\n") {
		if i == 0 || i == 1 {
			continue
		} else if len(v) < 1 {
			continue
		}
		linesList = append(linesList, v)
	}

	for _, v := range linesList {
		v = strings.TrimSpace(v)
		diskInfoTemp := DiskInfo{}
		gstatInfoList := reSplitSpace.Split(v, -1)

		diskInfoTemp.queueLength, _ = strconv.Atoi(gstatInfoList[0])
		diskInfoTemp.opsPerSec, _ = strconv.Atoi(gstatInfoList[1])
		diskInfoTemp.readsPerSecOp, _ = strconv.Atoi(gstatInfoList[2])
		diskInfoTemp.readsPerSecKb, _ = strconv.Atoi(gstatInfoList[3])
		readsTimePerOp, _ := strconv.ParseFloat(gstatInfoList[4], 32)
		diskInfoTemp.readsTimePerOp = float32(readsTimePerOp)

		diskInfoTemp.writesPerSecOp, _ = strconv.Atoi(gstatInfoList[5])
		diskInfoTemp.writesPerSecKb, _ = strconv.Atoi(gstatInfoList[6])
		writesTimePerOp, _ := strconv.ParseFloat(gstatInfoList[7], 32)
		diskInfoTemp.writesTimePerOp = float32(writesTimePerOp)

		busyPercent, _ := strconv.ParseFloat(gstatInfoList[8], 32)
		diskInfoTemp.busyPercent = float32(busyPercent)
		diskInfoTemp.deviceName = gstatInfoList[9]

		diskInfo = append(diskInfo, diskInfoTemp)
	}

	result := ""
	for i, v := range diskInfo {
		if i != 0 {
			result = result + "\n"
			_ = result // Static checker shits it's pants on the line above, that's why this is here
		}
		result = "gstat{disk=\"" + v.deviceName + "\",info=\"queue_length\"} " + strconv.Itoa(diskInfo[0].queueLength) + "\n"
		result = result + "gstat{disk=\"" + v.deviceName + "\",info=\"operations_per_second\"} " + strconv.Itoa(v.opsPerSec) + "\n"
		result = result + "gstat{disk=\"" + v.deviceName + "\",info=\"reads_per_second_ops\"} " + strconv.Itoa(v.readsPerSecOp) + "\n"
		result = result + "gstat{disk=\"" + v.deviceName + "\",info=\"reads_per_second_kbs\"} " + strconv.Itoa(v.readsPerSecKb) + "\n"
		readsTimePerOp := fmt.Sprintf("%.1f", v.readsTimePerOp)
		if readsTimePerOp == "0.0" {
			readsTimePerOp = "0"
		}
		result = result + "gstat{disk=\"" + v.deviceName + "\",info=\"reads_time_per_op\"} " + readsTimePerOp + "\n"
		result = result + "gstat{disk=\"" + v.deviceName + "\",info=\"writes_per_second_ops\"} " + strconv.Itoa(v.writesPerSecOp) + "\n"
		result = result + "gstat{disk=\"" + v.deviceName + "\",info=\"writes_per_second_kbs\"} " + strconv.Itoa(v.writesPerSecKb) + "\n"
		writesTimePerOp := fmt.Sprintf("%.1f", v.writesTimePerOp)
		if writesTimePerOp == "0.0" {
			writesTimePerOp = "0"
		}
		result = result + "gstat{disk=\"" + v.deviceName + "\",info=\"writes_time_per_op\"} " + writesTimePerOp + "\n"
		busyPercent := fmt.Sprintf("%.1f", v.busyPercent)
		if busyPercent == "0.0" {
			busyPercent = "0"
		}
		result = result + "gstat{disk=\"" + v.deviceName + "\",info=\"busy_percent\"} " + busyPercent
	}

	if !endsWithNewline.MatchString(result) {
		result = result + "\n"
	}
	return result
}

func getNodeExporterMetrics() string {
	url := "http://localhost:9100/metrics"
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
	if !endsWithNewline.MatchString(metricsString) {
		metricsString = metricsString + "\n"
	}
	return metricsString
}

func concatMetrics() string {
	var metricsTextFinal string
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		metricsText := getGstatMetrics()
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

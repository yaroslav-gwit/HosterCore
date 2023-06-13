package main

import (
	"fmt"
	"hoster/cmd"
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
		w.Header().Set("Content-Type", "text/plain; version=1")
		fmt.Fprint(w, "hoster metrics exporter\ngo to /metrics to see the exported metrics")
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metricsText := concatMetrics()
		w.Header().Set("Content-Type", "text/plain; version=1")
		fmt.Fprint(w, metricsText)
		clearMetricsList()
	})

	http.ListenAndServe(":9101", nil)
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

	wg.Add(1)
	go func() {
		metricsText := getZpoolInfo()
		addMetricsToList(metricsText)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		metricsText := getVmNumbers()
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
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("PANIC AVOIDED! Recovered: ", r)
		}
	}()

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

	result := "# HELP gstat A parsed output from the FreeBSD utility gstat for disk IO monitoring.\n"
	result = result + "# TYPE gstat gauge\n"
	for i, v := range diskInfo {
		// if i != 0 {
		// 	result = result + "\n"
		// 	_ = result // Static checker shits it's pants on the line above, that's why this is here
		// }
		_ = i
		result = result + "gstat{disk=\"" + v.deviceName + "\",info=\"queue_length\"} " + strconv.Itoa(diskInfo[0].queueLength) + "\n"
		result = result + "gstat{disk=\"" + v.deviceName + "\",info=\"operations_per_second\"} " + strconv.Itoa(v.opsPerSec) + "\n"
		result = result + "gstat{disk=\"" + v.deviceName + "\",info=\"reads_ops_per_second\"} " + strconv.Itoa(v.readsPerSecOp) + "\n"
		result = result + "gstat{disk=\"" + v.deviceName + "\",info=\"reads_kbs_per_second\"} " + strconv.Itoa(v.readsPerSecKb) + "\n"
		readsTimePerOp := fmt.Sprintf("%.1f", v.readsTimePerOp)
		if readsTimePerOp == "0.0" {
			readsTimePerOp = "0"
		}
		result = result + "gstat{disk=\"" + v.deviceName + "\",info=\"reads_time_per_op\"} " + readsTimePerOp + "\n"
		result = result + "gstat{disk=\"" + v.deviceName + "\",info=\"writes_ops_per_second\"} " + strconv.Itoa(v.writesPerSecOp) + "\n"
		result = result + "gstat{disk=\"" + v.deviceName + "\",info=\"writes_kbs_per_second\"} " + strconv.Itoa(v.writesPerSecKb) + "\n"
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
		result = result + "\n"
	}
	return result
}

type VmNumbers struct {
	all               int
	online            int
	backup            int
	offlineProduction int
}

func getVmNumbers() string {
	vmNumbers := VmNumbers{}
	vmNumbers.all, vmNumbers.online, vmNumbers.backup, vmNumbers.offlineProduction = cmd.VmNumbersOverview()

	result := "# HELP HosterCore related FreeBSD metrics.\n"
	result = result + "# TYPE hoster gauge\n"
	result = result + "hoster{counter=\"vms_all\"} " + strconv.Itoa(vmNumbers.all) + "\n"
	result = result + "hoster{counter=\"vms_online\"} " + strconv.Itoa(vmNumbers.online) + "\n"
	result = result + "hoster{counter=\"vms_backup\"} " + strconv.Itoa(vmNumbers.backup) + "\n"
	result = result + "hoster{counter=\"vms_offline_in_production\"} " + strconv.Itoa(vmNumbers.offlineProduction) + "\n"

	return result
}

type ZpoolInfo struct {
	name          string
	size          int64
	allocated     int64
	free          int64
	fragmentation int
	cap           int
	dedup         float64
	health        int
}

func getZpoolInfo() string {
	reSplitAtSpace := regexp.MustCompile(`\s+`)
	zpoolInfo := []ZpoolInfo{}
	out, err := exec.Command("zpool", "list", "-Hp").Output()
	if err != nil {
		log.Fatal("could not run zpool list: " + err.Error())
	}
	for _, v := range strings.Split(string(out), "\n") {
		if len(v) < 1 {
			continue
		}
		tempList := reSplitAtSpace.Split(v, -1)
		tempZpoolInfo := ZpoolInfo{}
		tempZpoolInfo.name = tempList[0]
		tempZpoolInfo.size, _ = strconv.ParseInt(tempList[1], 10, 64)
		tempZpoolInfo.allocated, _ = strconv.ParseInt(tempList[2], 10, 64)
		tempZpoolInfo.free, _ = strconv.ParseInt(tempList[3], 10, 64)
		tempZpoolInfo.fragmentation, _ = strconv.Atoi(tempList[6])
		tempZpoolInfo.cap, _ = strconv.Atoi(tempList[7])
		tempZpoolInfo.dedup, _ = strconv.ParseFloat(tempList[8], 64)
		if tempList[9] == "ONLINE" {
			tempZpoolInfo.health = 1
		}
		zpoolInfo = append(zpoolInfo, tempZpoolInfo)
	}
	result := ""
	for i, v := range zpoolInfo {
		if i != 0 {
			result = result + "\n"
			_ = result // Static checker shits it's pants on the line above, that's why this is here
		}
		result = "# HELP zpool_info A parsed output from the zpool info command under FreeBSD.\n"
		result = result + "# TYPE zpool_info gauge\n"
		zpoolSize := fmt.Sprintf("%d", v.size)
		result = result + "zpool_info{pool=\"" + v.name + "\",info=\"size\"} " + zpoolSize + "\n"
		zpoolAllocated := fmt.Sprintf("%d", v.allocated)
		result = result + "zpool_info{pool=\"" + v.name + "\",info=\"allocated\"} " + zpoolAllocated + "\n"
		zpoolFree := fmt.Sprintf("%d", v.free)
		result = result + "zpool_info{pool=\"" + v.name + "\",info=\"free\"} " + zpoolFree + "\n"
		result = result + "zpool_info{pool=\"" + v.name + "\",info=\"fragmentation\"} " + strconv.Itoa(v.fragmentation) + "\n"
		result = result + "zpool_info{pool=\"" + v.name + "\",info=\"cap\"} " + strconv.Itoa(v.cap) + "\n"
		zpoolDedup := fmt.Sprintf("%.2f", v.dedup)
		result = result + "zpool_info{pool=\"" + v.name + "\",info=\"dedup\"} " + zpoolDedup + "\n"
		result = result + "zpool_info{pool=\"" + v.name + "\",info=\"healthy\"} " + strconv.Itoa(v.health)
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
	return metricsString
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

package main

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"HosterCore/cmd"
	"HosterCore/internal/pkg/emojlog"
	FreeBSDLogger "HosterCore/internal/pkg/freebsd/logger"
	HosterHost "HosterCore/internal/pkg/hoster/host"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

// Global state vars
var vmInfoList []VmInfoStruct
var jailInfoList []JailInfoStruct

// var logChannel chan LogMessage
var upstreamServers []string

// Hardcoded failover DNS servers (in case user's main DNS server fails)
const DNS_SRV4_QUAD_NINE = "9.9.9.9:53"
const DNS_SRV4_CLOUD_FLARE = "1.1.1.1:53"

var log = logrus.New()

func init() {
	logFile := os.Getenv("LOG_FILE")

	// Log as JSON instead of the default ASCII/text formatter.
	// log.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)

	// Log to file, but fallback to STDOUT if something goes wrong
	if len(logFile) > 2 {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			FreeBSDLogger.LoggerToSyslog(FreeBSDLogger.LOGGER_SRV_SCHEDULER, FreeBSDLogger.LOGGER_LEVEL_ERROR, "DNS_SERVER: could not use this file for logging "+logFile+", falling back to STDOUT")
		} else {
			log.SetOutput(file)
		}
	}

	log.SetLevel(logrus.DebugLevel)
	// log.SetReportCaller(true)
}

func main() {
	// logFileOutput(LOG_SUPERVISOR, "Starting DNS server", logChannel)
	log.Info("Starting the DNS Server")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP)
	go func() {
		for sig := range signals {
			if sig == syscall.SIGHUP {
				// logFileOutput(LOG_SUPERVISOR, "Received a reload signal: SIGHUP", logChannel)
				log.Info("Received a reload signal: SIGHUP")
				vmInfoList = getVmsInfo()
				jailInfoList = getJailsInfo()
				loadUpstreamDnsServers()
			}
			if sig == syscall.SIGKILL {
				// logFileOutput(LOG_SUPERVISOR, "Received a stop signal: SIGKILL", logChannel)
				log.Info("Received a reload signal: SIGKILL")
				os.Exit(0)
			}
		}
	}()

	loadUpstreamDnsServers()

	vmInfoList = getVmsInfo()
	jailInfoList = getJailsInfo()

	server := dns.Server{Addr: ":53", Net: "udp"}
	server.Handler = dns.HandlerFunc(handleDNSRequest)

	// logFileOutput(LOG_SUPERVISOR, "DNS server listening on :53", logChannel)
	log.Info("DNS Server is listening on 0.0.0.0:53")
	err := server.ListenAndServe()
	if err != nil {
		emojlog.PrintLogMessage("Failed to start the DNS Server", emojlog.Error)
		os.Exit(1)
	}
}

// Parses and loads the list of upstream DNS servers from the host config file.
func loadUpstreamDnsServers() {
	// Load host config
	hostConf, err := HosterHost.GetHostConfig()
	if err != nil {
		log.Error("Error loading host config file:" + err.Error())
	}

	// Load upstream DNS servers from the host config
	upstreamServers = []string{}
	reMatchPort := regexp.MustCompile(`.*:\d+`)
	for _, v := range hostConf.DnsServers {
		if reMatchPort.MatchString(v) {
			upstreamServers = append(upstreamServers, v)
		} else {
			upstreamServers = append(upstreamServers, v+":53")
		}
	}

	// If host config doesn't include any servers, use the public ones
	if len(upstreamServers) < 1 {
		upstreamServers = append(upstreamServers, DNS_SRV4_QUAD_NINE)
		upstreamServers = append(upstreamServers, DNS_SRV4_CLOUD_FLARE)
	}

	log.Infof("Loaded these servers from the host config file: %s", upstreamServers)
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	var logLine string
	for _, q := range r.Question {
		clientIP := w.RemoteAddr().String()

		requestIsVmName := false
		requestIsJailName := false
		vmListIndex := 0
		jailListIndex := 0

		requestIsPublic := false
		dnsNameSplit := strings.Split(q.Name, ".")

		for i, v := range vmInfoList {
			dnsName := dnsNameSplit[0]
			if dnsName == v.vmName {
				requestIsVmName = true
				vmListIndex = i
			} else if dnsName == strings.ToLower(v.vmName) {
				requestIsVmName = true
				vmListIndex = i
			}
		}

		for i, v := range jailInfoList {
			dnsName := dnsNameSplit[0]
			if dnsName == v.JailName {
				requestIsJailName = true
				jailListIndex = i
			} else if dnsName == strings.ToLower(v.JailName) {
				requestIsJailName = true
				jailListIndex = i
			}
		}

		if len(dnsNameSplit) > 1 {
			// go func() { logFileOutput(LOG_DNS_GLOBAL, dnsNameSplit[len(dnsNameSplit)-2], logChannel) }()
			if IsPublicDomain(dnsNameSplit[len(dnsNameSplit)-2]) {
				requestIsPublic = true
			}
		}

		if requestIsPublic {
			response, server, err := queryExternalDNS(q)
			if err != nil {
				log.Error("Failed to query external DNS:", err)
				continue
			}
			m.Answer = append(m.Answer, response.Answer...)
			logLine = clientIP + " -> " + q.Name + "::." + parseAnswer(m.Answer) + " <- CACHE_MISS::" + server
			// go func() { logFileOutput(LOG_DNS_GLOBAL, logLine, logChannel) }()
			log.Info(logLine)
		} else if requestIsVmName {
			rr, err := dns.NewRR(q.Name + " IN A " + vmInfoList[vmListIndex].vmAddress)
			if err != nil {
				log.Error("Failed to create an A record:", err)
				continue
			}
			m.Answer = append(m.Answer, rr)
			logLine = clientIP + " -> " + q.Name + "::." + parseAnswer(m.Answer) + " <- CACHE_HIT::VM"
			// go func() { logFileOutput(LOG_DNS_LOCAL, logLine, logChannel) }()
			log.Info(logLine)
		} else if requestIsJailName {
			rr, err := dns.NewRR(q.Name + " IN A " + jailInfoList[jailListIndex].JailAddress)
			if err != nil {
				log.Error("Failed to create an A record:", err)
				continue
			}
			m.Answer = append(m.Answer, rr)
			logLine = clientIP + " -> " + q.Name + "::." + parseAnswer(m.Answer) + " <- CACHE_HIT::Jail"
			// go func() { logFileOutput(LOG_DNS_LOCAL, logLine, logChannel) }()
			log.Info(logLine)
		} else {
			response, server, err := queryExternalDNS(q)
			if err != nil {
				log.Error("Failed to query external DNS:", err)
				continue
			}
			m.Answer = append(m.Answer, response.Answer...)
			logLine = clientIP + " -> " + q.Name + "::." + parseAnswer(m.Answer) + " <- CACHE_MISS::" + server
			// go func() { logFileOutput(LOG_DNS_GLOBAL, logLine, logChannel) }()
			log.Info(logLine)
		}
	}

	err := w.WriteMsg(m)
	if err != nil {
		log.Error("Failed to send the DNS Response:" + err.Error())
	}
}

// Returns a DNS message, a server that returned the response, or an error
func queryExternalDNS(q dns.Question) (*dns.Msg, string, error) {
	c := dns.Client{}
	m := dns.Msg{}
	m.SetQuestion(q.Name, q.Qtype)

	var response *dns.Msg
	var err error
	var responseServer string

	// Try each DNS server until a response is received or all servers fail
	for _, server := range upstreamServers {
		response, _, err = c.Exchange(&m, server)
		if err == nil && response != nil && response.Rcode != dns.RcodeServerFailure {
			// Received a successful response, break the loop
			responseServer = server
			break
		}
	}

	if err != nil {
		return nil, "", err
	}

	return response, responseServer, nil
}

// Regex DNS Answer splitter
var reAnySpaceChar = regexp.MustCompile(`\s+`)

// Parses the DNS answer to only extract the IP address resolved
//
// Used purely for the logging purposes
func parseAnswer(msg []dns.RR) string {
	msgString := fmt.Sprintf("%s", msg)
	splitAnswer := reAnySpaceChar.Split(msgString, -1)
	result := ""
	for i, v := range splitAnswer {
		if i == len(splitAnswer)-1 {
			result = strings.Split(v, "]")[0]
		}
	}
	if result == "[" {
		result = "nil"
	}
	return result
}

type VmInfoStruct struct {
	vmName    string
	vmAddress string
}

func getVmsInfo() []VmInfoStruct {
	vmInfoVar := []VmInfoStruct{}
	allVms := cmd.GetAllVms()
	for _, v := range allVms {
		tempConfig := cmd.VmConfig(v)
		tempInfo := VmInfoStruct{}
		tempInfo.vmName = v
		tempInfo.vmAddress = tempConfig.Networks[0].IPAddress
		vmInfoVar = append(vmInfoVar, tempInfo)
	}
	return vmInfoVar
}

type JailInfoStruct struct {
	JailName    string
	JailAddress string
}

func getJailsInfo() (r []JailInfoStruct) {
	jails, err := HosterJailUtils.ListAllExtendedTable()
	if err != nil {
		return
	}

	for _, v := range jails {
		r = append(r, JailInfoStruct{JailName: v.Name, JailAddress: v.MainIpAddress})
	}
	return
}

const (
	LOG_SUPERVISOR = "supervisor"
	LOG_SYS_OUT    = "sys_stdout"
	LOG_SYS_ERR    = "sys_stderr"
	LOG_DNS_LOCAL  = "dns_locals"
	LOG_DNS_GLOBAL = "dns_global"
	LOG_DEV_DEBUG  = "dev_debug"
)

type LogMessage struct {
	Type    string
	Message string
}

func IsPublicDomain(topLevelDomain string) bool {
	for _, v := range publicDomainList {
		if strings.EqualFold(v, topLevelDomain) {
			// the above is the same as this:
			// if strings.ToUpper(v) == strings.ToUpper(topLevelDomain) {
			return true
		}
	}
	return false
}

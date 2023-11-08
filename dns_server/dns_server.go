package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"HosterCore/cmd"

	"github.com/miekg/dns"
)

// Global state vars
var vmInfoList []VmInfoStruct
var logChannel chan LogMessage
var upstreamServers []string

// Hardcoded failover DNS servers (in case user's main DNS server fails)
const DNS_SRV4_QUAD_NINE = "9.9.9.9:53"
const DNS_SRV4_CLOUD_FLARE = "1.1.1.1:53"

// Log file location
// const LOG_FILE_LOCATION = "/var/run/dns_server"  // OLD LOG
const LOG_FILE_LOCATION = "/var/log/hoster_dns_server.log"

func init() {
	logChannel = make(chan LogMessage)
	go startLogging(LOG_FILE_LOCATION, logChannel)
}

func main() {
	logFileOutput(LOG_SUPERVISOR, "Starting DNS server", logChannel)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP)
	go func() {
		for sig := range signals {
			if sig == syscall.SIGHUP {
				logFileOutput(LOG_SUPERVISOR, "Received a reload signal: SIGHUP", logChannel)
				vmInfoList = getVmsInfo()
				loadUpstreamDnsServers()
			}
			if sig == syscall.SIGKILL {
				logFileOutput(LOG_SUPERVISOR, "Received a stop signal: SIGKILL", logChannel)
				os.Exit(0)
			}
		}
	}()

	loadUpstreamDnsServers()

	vmInfoList = getVmsInfo()
	server := dns.Server{Addr: ":53", Net: "udp"}
	server.Handler = dns.HandlerFunc(handleDNSRequest)

	logFileOutput(LOG_SUPERVISOR, "DNS server listening on :53", logChannel)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Failed to start DNS server:", err)
	}
}

func loadUpstreamDnsServers() {
	hostConfig, err := cmd.GetHostConfig()
	if err != nil {
		logFileOutput(LOG_SUPERVISOR, "Error loading host config file: "+err.Error(), logChannel)
	}

	upstreamServers = []string{}
	reMatchPort := regexp.MustCompile(`.*:\d+`)
	for _, v := range hostConfig.DnsServers {
		if reMatchPort.MatchString(v) {
			upstreamServers = append(upstreamServers, v)
		} else {
			upstreamServers = append(upstreamServers, v+":53")
		}
	}
	debugText := fmt.Sprintf("Loaded these servers from the host config file: %s", upstreamServers)
	logFileOutput(LOG_SUPERVISOR, debugText, logChannel)
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	var logLine string
	for _, q := range r.Question {
		clientIP := w.RemoteAddr().String()
		requestIsVmName := false
		requestIsPublic := false
		vmListIndex := 0

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

		if len(dnsNameSplit) > 1 {
			// go func() { logFileOutput(LOG_DNS_GLOBAL, dnsNameSplit[len(dnsNameSplit)-2], logChannel) }()
			if IsPublicDomain(dnsNameSplit[len(dnsNameSplit)-2]) {
				requestIsPublic = true
			}
		}

		if requestIsPublic {
			response, server, err := queryExternalDNS(q)
			if err != nil {
				fmt.Println("Failed to query external DNS:", err)
				continue
			}
			m.Answer = append(m.Answer, response.Answer...)
			logLine = clientIP + "  ->  " + q.Name + "  <->  " + parseAnswer(m.Answer) + "  <-  " + server
			go func() { logFileOutput(LOG_DNS_GLOBAL, logLine, logChannel) }()
		} else if requestIsVmName {
			rr, err := dns.NewRR(q.Name + " IN A " + vmInfoList[vmListIndex].vmAddress)
			if err != nil {
				fmt.Println("Failed to create A record:", err)
				continue
			}
			m.Answer = append(m.Answer, rr)
			logLine = clientIP + "  ->  " + q.Name + "  <->  " + parseAnswer(m.Answer) + "  <-  local DB"
			go func() { logFileOutput(LOG_DNS_LOCAL, logLine, logChannel) }()
		} else {
			response, server, err := queryExternalDNS(q)
			if err != nil {
				fmt.Println("Failed to query external DNS:", err)
				continue
			}
			m.Answer = append(m.Answer, response.Answer...)
			logLine = clientIP + "  ->  " + q.Name + "  <->  " + parseAnswer(m.Answer) + "  <-  " + server + " (req is not public, nor the VM)"
			go func() { logFileOutput(LOG_DNS_GLOBAL, logLine, logChannel) }()
		}
	}

	err := w.WriteMsg(m)
	if err != nil {
		fmt.Println("Failed to send DNS response:", err)
	}
}

// Returns a DNS message, a server that returned the response, or an error
func queryExternalDNS(q dns.Question) (*dns.Msg, string, error) {
	c := dns.Client{}
	m := dns.Msg{}
	m.SetQuestion(q.Name, q.Qtype)

	// Set the list of DNS servers to try
	servers := upstreamServers
	servers = append(servers, DNS_SRV4_QUAD_NINE)
	servers = append(servers, DNS_SRV4_CLOUD_FLARE)

	var response *dns.Msg
	var err error
	var responseServer string

	// Try each DNS server until a response is received or all servers fail
	for _, server := range servers {
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
		result = "EMPTY RESPONSE"
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

func logFileOutput(msgType string, msgString string, logChannel chan LogMessage) {
	logChannel <- LogMessage{
		Type:    msgType,
		Message: msgString,
	}
}

func startLogging(logFileLocation string, logChannel chan LogMessage) {
	logFile, err := os.OpenFile(logFileLocation, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		_ = exec.Command("logger", err.Error()).Run()
	}

	defer func() {
		if r := recover(); r != nil {
			logFile, err = os.OpenFile(logFileLocation, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
			if err != nil {
				_ = exec.Command("logger", err.Error()).Run()
			}
			errorValue := fmt.Sprintf("PANIC AVOIDED: %v", r)
			_ = exec.Command("logger", errorValue).Run()
		}

		logFile.Close()
	}()

	for logMsg := range logChannel {
		timeNow := time.Now().Format("2006-01-02_15-04-05")
		logLine := timeNow + " [" + logMsg.Type + "] " + logMsg.Message + "\n"
		_, err := logFile.WriteString(logLine)
		if err != nil {
			_ = exec.Command("logger", err.Error()).Run()
		}
	}
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

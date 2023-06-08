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

	"hoster/cmd"

	"github.com/miekg/dns"
)

// Global state vars
var vmInfoList []VmInfoStruct
var logChannel chan LogMessage
var upstreamServers []string

func init() {
	logChannel = make(chan LogMessage)
	go startLogging("/var/run/dns_server", logChannel)
	loadUpstreamDnsServers()
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
	upstreamServers = append(upstreamServers, hostConfig.DnsServers...)
	reMatchPort := regexp.MustCompile(`.*:\d+`)
	for i, v := range upstreamServers {
		if reMatchPort.MatchString(v) {
			continue
		} else {
			upstreamServers[i] = v + ":53"
		}
	}
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	var logLine string
	for _, q := range r.Question {
		clientIP := w.RemoteAddr().String()
		requestIsVmName := false
		vmListIndex := 0
		for i, v := range vmInfoList {
			dnsName := strings.Split(q.Name, ".")[0]
			if dnsName == v.vmName {
				requestIsVmName = true
				vmListIndex = i
			} else if dnsName == strings.ToLower(v.vmName) {
				requestIsVmName = true
				vmListIndex = i
			}
		}

		if requestIsVmName {
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
			logLine = clientIP + "  ->  " + q.Name + "  <->  " + parseAnswer(m.Answer) + "  <-  " + server
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
	fmt.Println(servers)
	servers = append(servers, "9.9.9.9:53")
	servers = append(servers, "1.1.1.1:53")

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
	defer logFile.Close()

	for logMsg := range logChannel {
		timeNow := time.Now().Format("2006-01-02_15-04-05")
		logLine := timeNow + " [" + logMsg.Type + "] " + logMsg.Message + "\n"
		_, err := logFile.WriteString(logLine)
		if err != nil {
			_ = exec.Command("logger", err.Error()).Run()
		}
	}
}

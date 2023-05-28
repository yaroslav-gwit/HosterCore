package main

import (
	"fmt"
	"os"
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

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP)

	vmInfoList = getVmsInfo()
	fmt.Println(vmInfoList)
	fmt.Println()
	fmt.Println()

	server := dns.Server{Addr: ":53", Net: "udp"}
	server.Handler = dns.HandlerFunc(handleDNSRequest)

	fmt.Println("DNS server listening on :53")
	fmt.Println()
	fmt.Println()

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Failed to start DNS server:", err)
	}

	go func() {
		for sig := range signals {
			if sig == syscall.SIGHUP {
				fmt.Println()
				fmt.Println()
				fmt.Println("Received a reload signal: " + sig.String())
				vmInfoList = getVmsInfo()
				fmt.Println("New VM list:")
				fmt.Println(vmInfoList)
				fmt.Println()
				fmt.Println()
			}
		}
	}()
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	for _, q := range r.Question {
		clientIP := w.RemoteAddr().String()
		reqTime := time.Now().Format("2006-01-02_15:04:05")

		requestIsVmName := false
		vmListIndex := 0
		for i, v := range vmInfoList {
			dnsName := strings.Split(q.Name, ".")[0]
			if dnsName == v.vmName {
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
			fmt.Println(reqTime + "  " + clientIP + "  ->  " + q.Name + "  <->  " + parseAnswer(m.Answer) + " (from local DB)")
		} else {
			response, server, err := queryExternalDNS(q)
			if err != nil {
				fmt.Println("Failed to query external DNS:", err)
				continue
			}
			// for _, rr := range response.Answer {
			// 	m.Answer = append(m.Answer, rr)
			// }
			m.Answer = append(m.Answer, response.Answer...)
			fmt.Println(reqTime + "  " + clientIP + "  ->  " + q.Name + "  <->  " + parseAnswer(m.Answer) + " (from server: " + server + ")")
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
	servers := []string{"9.9.9.9:53", "1.1.1.1:53"}

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

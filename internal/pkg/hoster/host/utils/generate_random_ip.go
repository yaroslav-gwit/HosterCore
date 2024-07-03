package HosterHostUtils

import (
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"slices"
	"time"
)

func GenerateNewRandomIp(networkName string) (r string, e error) {
	var existingIps []string

	// Add existing VM IPs
	vms, err := HosterVmUtils.ListJsonApi()
	if err != nil {
		e = err
		return
	}

	for _, v := range vms {
		for _, vv := range v.Networks {
			if vv.NetworkBridge == networkName {
				existingIps = append(existingIps, vv.IPAddress)
			}
		}
	}
	// EOF Add existing VM IPs

	// Add existing Jail IPs
	jailList, err := HosterJailUtils.ListJsonApi()
	if err != nil {
		e = err
		return
	}
	for _, v := range jailList {
		existingIps = append(existingIps, v.IPAddress)
	}
	// EOF Add existing Jail IPs

	networks, err := HosterNetwork.GetNetworkConfig()
	if err != nil {
		e = errors.New("could not read the config file: " + err.Error())
		return
	}

	var subnet string
	var rangeStart string
	var rangeEnd string
	networkNameFoundGlobal := false
	for i, v := range networks {
		if v.NetworkName == networkName {
			networkNameFoundGlobal = true
			subnet = networks[i].Subnet
			rangeStart = networks[i].RangeStart
			rangeEnd = networks[i].RangeEnd
		}
	}
	if !networkNameFoundGlobal {
		subnet = networks[0].Subnet
		rangeStart = networks[0].RangeStart
		rangeEnd = networks[0].RangeEnd
	}

	var randomIp string
	randomIp, err = generateUniqueRandomIp(subnet)
	if err != nil {
		e = errors.New("could not generate a random IP address: " + err.Error())
		return
	}

	iteration := 0
	for {
		if slices.Contains(existingIps, randomIp) || !ipIsWithinRange(randomIp, subnet, rangeStart, rangeEnd) {
			randomIp, err = generateUniqueRandomIp(subnet)
			if err != nil {
				e = errors.New("could not generate a random IP address: " + err.Error())
				return
			}

			iteration++
			if iteration > 800 {
				e = errors.New("ran out of IP available addresses within this range")
				return
			}
		} else {
			r = randomIp
			return
		}
	}
}

func generateUniqueRandomIp(subnet string) (string, error) {
	// Set the seed for the random number generator
	rand.Seed(time.Now().UnixNano())

	// Parse the subnet IP and mask
	ip, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", errors.New(err.Error())
	}

	// Calculate the size of the address space within the subnet
	size, _ := ipNet.Mask.Size()
	numHosts := (1 << (32 - size)) - 2

	// Generate a random host address within the subnet
	host := rand.Intn(numHosts) + 1
	addr := ip.Mask(ipNet.Mask)
	addr[0] |= byte(host >> 24)
	addr[1] |= byte(host >> 16)
	addr[2] |= byte(host >> 8)
	addr[3] |= byte(host)

	stringAddress := fmt.Sprintf("%v", addr)
	return stringAddress, nil
}

func ipIsWithinRange(ipAddress string, subnet string, rangeStart string, rangeEnd string) bool {
	// Parse the subnet IP and mask
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		panic(err)
	}

	// Define the range of allowed host addresses
	start := net.ParseIP(rangeStart).To4()
	end := net.ParseIP(rangeEnd).To4()

	// Parse the IP address to check
	ip := net.ParseIP(ipAddress).To4()

	// Check if the IP address is within the allowed range
	if ipNet.Contains(ip) && bytesInRange(ip, start, end) {
		return true
	} else {
		return false
	}
}

func bytesInRange(ip, start, end []byte) bool {
	for i := 0; i < len(ip); i++ {
		if start[i] > end[i] {
			log.Fatal("Make sure range start is lower than range end!")
		} else if ip[i] < start[i] || ip[i] > end[i] {
			return false
		}
	}
	return true
}

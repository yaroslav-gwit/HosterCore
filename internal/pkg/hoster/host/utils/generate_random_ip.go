package HosterHostUtils

import (
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"golang.org/x/exp/slices"
)

func GenerateNewRandomIp(networkName string) (string, error) {
	var existingIps []string

	// TO DO
	// Add existing VM IPs
	// for _, v := range getAllVms() {
	// 	networkNameFound := false
	// 	tempConfig := vmConfig(v)
	// 	for i, v := range tempConfig.Networks {
	// 		if v.NetworkBridge == networkName {
	// 			networkNameFound = true
	// 			existingIps = append(existingIps, tempConfig.Networks[i].IPAddress)
	// 		}
	// 	}
	// 	if !networkNameFound {
	// 		existingIps = append(existingIps, tempConfig.Networks[0].IPAddress)
	// 	}
	// }
	// EOF Add existing VM IPs

	// Add existing Jail IPs
	jailList, err := HosterJailUtils.ListAllExtendedTable()
	if err != nil {
		return "", err
	}
	for _, v := range jailList {
		existingIps = append(existingIps, v.MainIpAddress)
	}
	// EOF Add existing Jail IPs

	networks, err := HosterNetwork.GetNetworkConfig()
	if err != nil {
		return "", errors.New("could not read the config file: " + err.Error())
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
	// var err error
	randomIp, err = generateUniqueRandomIp(subnet)
	if err != nil {
		return "", errors.New("could not generate a random IP address: " + err.Error())
	}

	iteration := 0
	for {
		if slices.Contains(existingIps, randomIp) || !ipIsWithinRange(randomIp, subnet, rangeStart, rangeEnd) {
			randomIp, err = generateUniqueRandomIp(subnet)
			if err != nil {
				return "", errors.New("could not generate a random IP address: " + err.Error())
			}
			iteration = iteration + 1
			if iteration > 400 {
				return "", errors.New("ran out of IP available addresses within this range")
			}
		} else {
			break
		}
	}

	return randomIp, nil
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

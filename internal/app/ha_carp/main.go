package main

import (
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	"fmt"
	"net"
	"os"
	"time"
)

var version = "" // version is set by the build system

func main() {
	// Print the version and exit
	args := os.Args
	if len(args) > 1 {
		res := os.Args[1]
		if res == "version" || res == "v" || res == "--version" || res == "-v" {
			fmt.Println(version)
			return
		}
	}

	// Remove the old socket if it exists
	if _, err := os.Stat(CarpUtils.SOCKET_FILE); err == nil {
		os.Remove(CarpUtils.SOCKET_FILE)
	}

	// Create the Unix socket listener
	listener, err := net.Listen("unix", CarpUtils.SOCKET_FILE)
	if err != nil {
		log.Fatalf("Error creating Unix socket: %v", err)
	}

	// Clean up the socket file and listener
	defer log.Info("HA Module is shutting down")
	defer os.Remove(CarpUtils.SOCKET_FILE)
	defer listener.Close()

	log.Infof("HA Module has started listening on %s", CarpUtils.SOCKET_FILE)

	go func() { // Check if the current node is the master
		for {
			checkIfMaster()
			time.Sleep(selfMasterCheckInterval)
		}
	}()

	go func() { // Ping master in order to refresh our online status
		for {
			pingMaster()
			time.Sleep(pingRemoteMasterInterval)
		}
	}()

	go func() { // Detect offline hosts
		for {
			detectOfflineHosts()
			time.Sleep(21 * time.Second)
		}
	}()

	go func() { // Failover resources
		for {
			failOverResource()
			time.Sleep(6 * time.Second)
		}
	}()

	go func() { // Sync local master state with other nodes
		for {
			syncState()
			time.Sleep(120 * time.Second)
		}
	}()

	go func() { // Sync local master state with other nodes
		for {
			getRemoteBackups()
			// time.Sleep(15 * time.Minute)
			time.Sleep(3 * time.Minute) // Temporary short-time value for testing
		}
	}()

	for { // Accept incoming socket connections
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Error accepting connection:", err)
			continue
		}

		// Handle the connection in a new goroutine
		go handleConnection(conn)
	}
}

func isSelfMaster() bool {
	if !iAmMaster {
		return false
	}

	if time.Now().Local().Unix()-becameMaster < 25 { // Don't add offline backups for the first 15 seconds after becoming master
		return false
	}

	return true
}

func addOfflineBackup(hostname string) {
	if !isSelfMaster() {
		return
	}

	getRemoteBackups()

	mutexOfflineBackups.Lock()
	mutexBackups.RLock()

	for _, v := range backups {
		if v.ParentHost == hostname {
			offlineBackups = append(offlineBackups, v)
			log.Warnf("Added backup %s to offline backups", v.ResourceName)
		}
	}

	mutexBackups.RUnlock()
	mutexOfflineBackups.Unlock()
}

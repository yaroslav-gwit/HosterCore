package main

import (
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	"bufio"
	"encoding/json"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Create a new reader
	reader := bufio.NewReader(conn)
	// Read the message until a newline character is encountered
	messageBytes, err := reader.ReadBytes('\n')
	if err != nil {
		log.Println("Error reading message:", err)
		return
	}
	// Remove the newline before processing
	messageBytes = messageBytes[:len(messageBytes)-1]

	// First, unmarshal the base struct to detect the payload type
	var base CarpUtils.BasePayload
	err = json.Unmarshal(messageBytes, &base)
	if err != nil {
		log.Error("Error unmarshalling JSON:", err)
		return
	}

	// Switch on the type field to unmarshal into the correct struct
	switch base.Type {
	case "host_add":
		var payload CarpUtils.HostInfo
		err = json.Unmarshal(messageBytes, &payload)
		if err != nil {
			log.Error("Error unmarshalling Hosts Payload:", err)
			return
		}
		// addNewHost(payload)
		receivePing(payload)
		log.Debugf("Received Hosts Payload: %+v", payload)
		closeWithSuccess(conn)

	case "host_list":
		var payload CarpUtils.HostInfo
		err = json.Unmarshal(messageBytes, &payload)
		if err != nil {
			log.Error("Error unmarshalling Hosts Payload:", err)
			return
		}

		resp := listHosts()
		// Send a response back
		respBytes, _ := json.Marshal(resp)
		respBytes = append(respBytes, []byte("\n")...)
		conn.Write(respBytes)
		log.Debugf("Received Hosts Payload: %+v", payload)
		return

	case "backup_add":
		var payload CarpUtils.BackupInfo
		err = json.Unmarshal(messageBytes, &payload)
		if err != nil {
			log.Error("Error unmarshalling Backups Payload:", err)
			return
		}

		addNewBackup(payload)
		log.Debugf("Received Backups Payload: %+v", payload)
		closeWithSuccess(conn)

	case "backup_list":
		var payload CarpUtils.HostInfo
		err = json.Unmarshal(messageBytes, &payload)
		if err != nil {
			log.Error("Error unmarshalling Hosts Payload:", err)
			return
		}

		resp := listBackups()
		// Send a response back
		respBytes, _ := json.Marshal(resp)
		respBytes = append(respBytes, []byte("\n")...)
		conn.Write(respBytes)
		log.Debugf("Received Hosts Payload: %+v", payload)
		return

	case "ha_status":
		ha := CarpUtils.HaStatus{}
		if iAmMaster {
			ha.Status = "MASTER"
		} else {
			ha.Status = "FOLLOWER"
		}
		ha.Hosts = listHosts()
		ha.Resources = listBackups()
		ha.CurrentMaster = currentMaster
		ha.ServiceHealth = "OK"

		// Send a response back
		respBytes, err := json.Marshal(ha)
		if err != nil {
			log.Error("Error marshalling ha_status response:", err)
			log.Debugf("Response dump: %+v", ha)
			closeWithFailure(conn)
			return
		}

		respBytes = append(respBytes, []byte("\n")...)
		conn.Write(respBytes)
		log.Debugf("Responded: %+v", ha)
		return

	case "ha_receive_hosts":
		if iAmMaster {
			log.Debug("Received state from a follower, not syncing")
			closeWithSuccess(conn)
			return
		}

		info := CarpUtils.HaStatus{}
		err = json.Unmarshal(messageBytes, &info)
		if err != nil {
			log.Warn("Unknown payload type:", base.Type)
			closeWithFailure(conn)
		}

		mutexHosts.Lock()
		hosts = []CarpUtils.HostInfo{}
		hosts = append(hosts, info.Hosts...)
		mutexHosts.Unlock()

		mutexBackups.Lock()
		backups = []CarpUtils.BackupInfo{}
		backups = append(backups, info.Resources...)
		mutexBackups.Unlock()

		closeWithSuccess(conn)
		log.Debug("Received and synced state from master")
		return

	default:
		log.Warn("Unknown payload type:", base.Type)
		closeWithFailure(conn)
		return
	}
}

func closeWithSuccess(conn net.Conn) {
	response := CarpUtils.SocketResponse{
		Success: true,
	}
	respBytes, _ := json.Marshal(response)
	respBytes = append(respBytes, []byte("\n")...)
	conn.Write(respBytes)
}

func closeWithFailure(conn net.Conn) {
	response := CarpUtils.SocketResponse{
		Success: true,
	}
	respBytes, _ := json.Marshal(response)
	respBytes = append(respBytes, []byte("\n")...)
	conn.Write(respBytes)
}

package main

import (
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	"encoding/json"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	var buf [512]byte
	n, err := conn.Read(buf[:])
	if err != nil {
		log.Error("Error reading from connection:", err)
		return
	}

	// First, unmarshal the base struct to detect the payload type
	var base CarpUtils.BasePayload
	err = json.Unmarshal(buf[:n], &base)
	if err != nil {
		log.Error("Error unmarshalling JSON:", err)
		return
	}

	// Switch on the type field to unmarshal into the correct struct
	switch base.Type {
	case "host_add":
		var payload CarpUtils.HostInfo
		err = json.Unmarshal(buf[:n], &payload)
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
		err = json.Unmarshal(buf[:n], &payload)
		if err != nil {
			log.Error("Error unmarshalling Hosts Payload:", err)
			return
		}

		resp := listHosts()
		// Send a response back
		respBytes, _ := json.Marshal(resp)
		conn.Write(respBytes)
		log.Debugf("Received Hosts Payload: %+v", payload)
		return

	case "backup_add":
		var payload CarpUtils.BackupInfo
		err = json.Unmarshal(buf[:n], &payload)
		if err != nil {
			log.Error("Error unmarshalling Backups Payload:", err)
			return
		}

		addNewBackup(payload)
		log.Debugf("Received Backups Payload: %+v", payload)
		closeWithSuccess(conn)

	case "backup_list":
		var payload CarpUtils.HostInfo
		err = json.Unmarshal(buf[:n], &payload)
		if err != nil {
			log.Error("Error unmarshalling Hosts Payload:", err)
			return
		}

		resp := listBackups()
		// Send a response back
		respBytes, _ := json.Marshal(resp)
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

		// Send a response back
		respBytes, _ := json.Marshal(ha)
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
		err := json.Unmarshal(buf[:n], &info)
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
	conn.Write(respBytes)
}

func closeWithFailure(conn net.Conn) {
	response := CarpUtils.SocketResponse{
		Success: true,
	}
	respBytes, _ := json.Marshal(response)
	conn.Write(respBytes)
}

package CarpClient

import (
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	"encoding/json"
	"fmt"
	"net"
)

func HostAdd(input CarpUtils.HostInfo) error {
	conn, err := net.Dial("unix", CarpUtils.SOCKET_FILE)
	if err != nil {
		return fmt.Errorf("can't connect to Unix socket: " + err.Error())
	}
	defer conn.Close()

	// Marshal the payload to JSON
	input.Type = "host_add"
	payloadBytes, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Send the JSON payload to the server
	_, err = conn.Write(payloadBytes)
	if err != nil {
		return fmt.Errorf("error sending data: %v", err)
	}

	// Read the server's response
	var buf [512]byte
	_, err = conn.Read(buf[:])
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	return nil
}

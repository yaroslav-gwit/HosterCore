package CarpClient

import (
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

func ReceiveHostAdd(input CarpUtils.HostInfo) error {
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
	// _, err = conn.Write(payloadBytes)
	// if err != nil {
	// 	return fmt.Errorf("error sending data: %v", err)
	// }
	// Send the JSON payload to the server
	jsonDataWithNewLine := append(payloadBytes, []byte("\n")...)
	_, err = conn.Write(jsonDataWithNewLine)
	if err != nil {
		return fmt.Errorf("error sending data: %v", err)
	}

	// Read the server's response
	reader := bufio.NewReader(conn)
	// Read the message until a newline character is encountered
	messageBytes, err := reader.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("error reading message: %v", err)
	}
	// Remove the newline before processing
	messageBytes = messageBytes[:len(messageBytes)-1]
	_ = messageBytes // discard the response

	return nil
}

func GetHaStatus() (r CarpUtils.HaStatus, e error) {
	conn, err := net.Dial("unix", CarpUtils.SOCKET_FILE)
	if err != nil {
		e = fmt.Errorf("can't connect to Unix socket: " + err.Error())
		return
	}
	defer conn.Close()

	// Marshal the payload to JSON
	input := CarpUtils.BasePayload{}
	input.Type = "ha_status"
	payloadBytes, err := json.Marshal(input)
	if err != nil {
		e = fmt.Errorf("error marshaling JSON: %v", err)
		return
	}

	// Send the JSON payload to the server
	jsonDataWithNewLine := append(payloadBytes, []byte("\n")...)
	_, err = conn.Write(jsonDataWithNewLine)
	if err != nil {
		e = fmt.Errorf("error sending data: %v", err)
		return
	}

	// Read and decode the server's response
	reader := bufio.NewReader(conn)
	// Read the message until a newline character is encountered
	messageBytes, err := reader.ReadBytes('\n')
	if err != nil {
		e = fmt.Errorf("error reading message: %v", err)
		return
	}
	// Remove the newline before processing
	messageBytes = messageBytes[:len(messageBytes)-1]

	// Read the server's response
	// var buf [5120000]byte
	// out, err := conn.Read(buf[:])
	// if err != nil {
	// 	e = fmt.Errorf("error reading response: %v", err)
	// 	return
	// }

	// Unmarshal the response
	err = json.Unmarshal(messageBytes, &r)
	if err != nil {
		e = fmt.Errorf("error unmarshaling JSON: %v", err)
		return
	}

	return
}

func ReceiveRemoteState(input CarpUtils.HaStatus) error {
	conn, err := net.Dial("unix", CarpUtils.SOCKET_FILE)
	if err != nil {
		return fmt.Errorf("can't connect to Unix socket: " + err.Error())
	}
	defer conn.Close()

	// Marshal the payload to JSON
	input.Type = "ha_receive_hosts"
	payloadBytes, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Send the JSON payload to the server
	jsonDataWithNewLine := append(payloadBytes, []byte("\n")...)
	_, err = conn.Write(jsonDataWithNewLine)
	if err != nil {
		return fmt.Errorf("error sending data: %v", err)
	}

	// Read the server's response
	reader := bufio.NewReader(conn)
	// Read the message until a newline character is encountered
	messageBytes, err := reader.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("error reading message: %v", err)
	}
	// Remove the newline before processing
	messageBytes = messageBytes[:len(messageBytes)-1]
	_ = messageBytes // discard the response

	return nil
}

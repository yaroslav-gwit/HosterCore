package main

import (
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

var jobs string
var sockAddr = "/tmp/echo.sock"

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go socketServer(&wg)

	wg.Wait()
}

func socketServer(wg *sync.WaitGroup) {
	if err := os.RemoveAll(sockAddr); err != nil {
		log.Fatal(err)
	}

	newSocket, err := net.Listen("unix", sockAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	defer wg.Done()
	defer newSocket.Close()

	for {
		conn, err := newSocket.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}

		go socketReceive(conn)
	}
}

func socketReceive(c net.Conn) {
	log.Printf("Client connected [%s]", c.RemoteAddr().Network())

	buffer := make([]byte, 0)
	dynamicBuffer := make([]byte, 1024)

	for {
		bytes, err := c.Read(dynamicBuffer)
		if err != nil {
			log.Printf("Error [%s]", err)
			break
		}

		buffer = append(buffer, dynamicBuffer[0:bytes]...)

		if bytes < len(dynamicBuffer) {
			break
		}
	}

	message := strings.TrimSuffix(string(buffer), "\n")
	log.Printf("Client has sent a message [%s]", message)

	c.Close()
}

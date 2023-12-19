package main

import (
	"log"
	"net"
	"os"
	"strings"
)

var jobs = []string{}
var sockAddr = "/tmp/echo.sock"

func main() {
	if err := os.RemoveAll(sockAddr); err != nil {
		log.Fatal(err)
	}

	l, err := net.Listen("unix", sockAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer l.Close()

	for {
		// Accept new connections, dispatching them to echoServer
		// in a goroutine.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}

		go echoServer(conn)
	}
}

func echoServer(c net.Conn) {
	log.Printf("Client connected [%s]", c.RemoteAddr().Network())
	// io.Copy(c, c)

	buffer := make([]byte, 0)
	tmpBuffer := make([]byte, 1024)

	for {
		bytes, err := c.Read(tmpBuffer)
		if err != nil {
			log.Printf("Error [%s]", err)
			break
		}

		buffer = append(buffer, tmpBuffer[0:bytes]...)

		if bytes < len(tmpBuffer) {
			break
		}
	}

	message := strings.TrimSuffix(string(buffer), "\n")
	log.Printf("Client has sent a message [%s]", message)

	// if strings.TrimSpace(string(buffer[0:bytes])) == "exit" || strings.TrimSpace(string(buffer[0:bytes])) == "end" {
	c.Close()
	// }
}

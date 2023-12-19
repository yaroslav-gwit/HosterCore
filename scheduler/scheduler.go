package main

import (
	"log"
	"net"
	"os"
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

	buffer := make([]byte, 1024)
	bytes, err := c.Read(buffer)
	if err != nil {
		log.Printf("Error [%v]", err)
	} else {
		log.Printf("Client has sent a message [%s]", buffer[:bytes])
	}

	c.Close()
}

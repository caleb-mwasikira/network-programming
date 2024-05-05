package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

var (
	host string = "0.0.0.0"
	port uint16 = 8080
)

func main() {
	addr := fmt.Sprintf("%v:%v", host, port)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("error connecting to server: %v", err)
	}
	defer conn.Close()

	// send message to server
	msg := "hello server"
	_, err = conn.Write([]byte(msg))
	if err != nil {
		log.Fatalf("error messaging server: %v", err)
	}
	fmt.Println("greeting server with: ", msg)

	// read data from the connection
	buffer := make([]byte, 1024)

	conn.SetReadDeadline(time.Now().Add(8 * time.Second))

	n, err := conn.Read(buffer)
	if err != nil {
		if err == io.EOF {
			fmt.Println("end of stream")
			return
		}

		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			fmt.Println("timeout occurred:", err)
			return
		}

		log.Fatalf("failed to read stream from connection; %v", err)
	}

	// print the received data
	fmt.Println("server says: ", string(buffer[:n]))
}

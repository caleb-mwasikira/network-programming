package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"path"
	"time"

	"caleb-mwasikira/network-programming/utils"
)

var (
	host               string        = "0.0.0.0"
	logger             *utils.Logger = utils.CreateLogger("server.log")
	min_port, max_port uint16        = 8000, 9000
)

func scanTCPPort(host string, port uint16) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, 30*time.Second)
	if err != nil {
		return false // connection refused - port closed
	}
	conn.Close()

	fmt.Printf("TCP address %v currently in use\n", address)
	return true // connection accepted - port open
}

// scans all ports on local machine within max_port and min_port
// and returns the address of the first closed port
func serverAddress(host string, port uint16) string {
	fmt.Println("scanning for closed ports...")

	if !scanTCPPort(host, port) {
		address := fmt.Sprintf("%v:%v", host, port)
		return address
	}

	for port := min_port + 1; port < max_port; port++ {
		if !scanTCPPort(host, port) {
			address := fmt.Sprintf("%v:%v", host, port)
			return address
		}
	}

	log.Fatalf("failed to start server! no available ports left within range (%v - %v)\n", min_port, max_port)
	return ""
}

func handleCLient(conn net.Conn) {
	defer conn.Close()
	client_address := conn.RemoteAddr().String()
	logger.Printf("new client connection from address %v", client_address)

	// read data from the connection
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	n, err := conn.Read(buffer)
	if err != nil {
		if err == io.EOF {
			fmt.Println("end of stream")
			return
		}

		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			logger.Println("timeout occurred:", err)
			return
		}

		return
	}
	msg := string(buffer[:n])
	logger.Printf("message received from client; %v\n", msg)

	// send message to client
	msg = "message received"
	logger.Printf("sending message to client; %v\n", msg)
	_, err = conn.Write([]byte(msg))
	if err != nil {
		logger.Fatalf("error sending message to client: %v", err)
	}

	logger.Printf("closing client connection %v\r\n\r\n", client_address)
}

func main() {
	// load SSL certificates
	cert_file := path.Join(utils.ProjectPath, "certs/server.crt")
	key_file := path.Join(utils.ProjectPath, "certs/server.key")

	cert, err := tls.LoadX509KeyPair(cert_file, key_file)
	if err != nil {
		logger.Fatalf("failed to setup secure communications; %v\n", err)
	}

	server_address := serverAddress(host, min_port)
	server, err := tls.Listen("tcp", server_address, &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		logger.Fatalf("failed to start server; %v", err)
	}
	defer server.Close()
	logger.Printf("TCP/SSL server started on address %v\n", server_address)

	for {
		logger.Printf("Waiting for client connections...")
		conn, err := server.Accept()
		if err != nil {
			logger.Printf("error accepting client %v", err)
		}

		go handleCLient(conn)
	}
}

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"time"

	"caleb-mwasikira/network-programming/utils"
)

var (
	server_address = "0.0.0.0:8000"
	logger         = utils.CreateLogger("client.log")
)

func main() {
	// load root certificate file
	root_cert := path.Join(utils.ProjectPath, "certs/ca-cert.pem")
	root_cert_data, err := os.ReadFile(root_cert)
	if err != nil {
		logger.Fatalf("failed to load root certificate; %v\n", err)
	}

	// connecting with a custom root certificate set
	cert_pool := x509.NewCertPool()
	ok := cert_pool.AppendCertsFromPEM(root_cert_data)
	if !ok {
		logger.Fatal("failed to parse root certificate")
	}

	logger.Printf("dialing server address %v...\n", server_address)
	conn, err := tls.Dial("tcp", server_address, &tls.Config{
		RootCAs: cert_pool,
	})
	if err != nil {
		logger.Fatalf("error connecting to server; %v", err)
	}
	defer conn.Close()

	// send message to server
	msg := "hello server"
	logger.Printf("sending message to server; %v\n", msg)
	_, err = conn.Write([]byte(msg))
	if err != nil {
		logger.Fatalf("error messaging server; %v", err)
	}

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
	msg = string(buffer[:n])
	logger.Printf("message received from server; %v\n", msg)
}

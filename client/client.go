package main

import (
	"caleb-mwasikira/network-programming/projectpath"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"time"
)

var (
	host string = "0.0.0.0"
	port uint16 = 8080
)

func main() {
	addr := fmt.Sprintf("%v:%v", host, port)

	// load root certificate file
	root_cert := path.Join(projectpath.Root, "certs/ca-cert.pem")
	root_cert_data, err := os.ReadFile(root_cert)
	if err != nil {
		log.Fatalf("failed to load root certificate; %v\n", err)
	}

	// connecting with a custom root certificate set
	cert_pool := x509.NewCertPool()
	ok := cert_pool.AppendCertsFromPEM(root_cert_data)
	if !ok {
		log.Fatal("failed to parse root certificate")
	}

	conn, err := tls.Dial("tcp", addr, &tls.Config{
		RootCAs: cert_pool,
	})
	if err != nil {
		log.Fatalf("error connecting to server: %v", err)
	}
	defer conn.Close()

	// send message to server
	_, err = conn.Write([]byte("hello server"))
	if err != nil {
		log.Fatalf("error messaging server: %v", err)
	}
	fmt.Println("sending message to server...")

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

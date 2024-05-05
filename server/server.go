package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"caleb-mwasikira/network-programming/projectpath"

	"github.com/google/uuid"
)

type Server struct {
	Id   string
	Host string
	Port uint16
	log  *log.Logger
}

func (s *Server) Addr() string {
	return net.JoinHostPort(s.Host, fmt.Sprint(s.Port))
}

func networkAddr() (string, error) {
	var ip_addr string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			ip_addr = ipNet.IP.String()
			break
		}
	}
	return ip_addr, nil
}

func generateServerId(id string) string {
	if len(id) != 0 {
		return id
	}

	id = uuid.NewString()
	fields := strings.Split(id, "-")
	return fields[len(fields)-1]
}

func testConnection(host string, port uint16) bool {
	addr := fmt.Sprintf("%v:%v", host, port)

	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		// connection failed
		return false
	}

	if conn != nil {
		// connection accepted
		defer conn.Close()
		return true
	}
	return false
}

func getFreePort(host string, min uint16, max uint16) uint16 {
	for port := min; port < max; port++ {
		addr := fmt.Sprintf("%v:%v", host, port)

		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err != nil {
			// connection refused - closed port (free to use)
			return port
		}

		if conn != nil {
			defer conn.Close()
			// connection accepted - open port (currently in use)
			continue
		}
	}
	return 0
}

func openLogFile(filename string) (*os.File, error) {
	log_dir := filepath.Join(projectpath.Root, ".logs/")

	err := os.MkdirAll(log_dir, 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	log_filepath := filepath.Join(log_dir, filename)
	file, err := os.OpenFile(log_filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	return file, nil

}

func NewServer(id string, host string, port uint16) *Server {
	// validate host
	if ip := net.ParseIP(host); ip == nil {
		fmt.Println("invalid IP address provided as host")

		// set host to network address, fallback on loopback address if err
		var err error = nil
		host, err = networkAddr()
		if err != nil {
			fmt.Printf("failed to set host as network address: %v\n", err)
			fmt.Println("falling back to loopback address...")
			host = "127.0.0.1"
		}
	}

	// validate port
	if ok := testConnection(host, port); ok {
		port = getFreePort(host, 1024, 49151)
		if port == 0 {
			log.Fatal("no closed ports available for server to run on")
		}
	}

	//
	server_id := generateServerId(id)
	var log_wrt io.Writer = os.Stdout

	file, err := openLogFile(fmt.Sprintf("%v.log", server_id))
	if err != nil {
		fmt.Printf("error saving logs to file: %v", err)
	} else {
		log_wrt = io.MultiWriter(os.Stdout, file)
	}

	return &Server{
		Id:   server_id,
		Host: host,
		Port: port,
		log:  log.New(log_wrt, "", log.LstdFlags|log.Lshortfile),
	}
}

func handleCLient(conn net.Conn) {
	defer conn.Close()

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
			fmt.Println("timeout occurred:", err)
			return
		}

		return
	}

	fmt.Println("client says: ", string(buffer[:n]))

	// send message to client
	_, err = conn.Write([]byte("message received"))
	if err != nil {
		fmt.Printf("error sending message to client: %v", err)
		return
	}
	fmt.Println("sending message to client...")
}

func main() {
	s := NewServer("", "0.0.0.0", 8080)
	listener, err := net.Listen("tcp", s.Addr())
	if err != nil {
		log.Fatalf("failed to start listener; %v", err)
	}
	defer listener.Close()
	s.log.Printf("listener started on port %v\nwaiting for client connections...", s.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.log.Printf("error accepting client %v", err)
		}
		s.log.Printf("new client connection to remote address %v", conn.RemoteAddr().String())

		go handleCLient(conn)
	}
}

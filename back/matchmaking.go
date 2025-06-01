package back

import (
	"fmt"
	"net"
	"sync"
)

type Client struct {
    Conn   net.Conn
    Pseudo string
}

var queue []Client
var mutex sync.Mutex


func StartTCPServer(){
	ln, err := net.Listen("tcp", ":9000") // port TCP 
	if err != nil {
		fmt.Println(" Erreur TCP", err)
		return
	}
	fmt.Println(" TCP server is running on port 9000")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go HandleClient(conn)
	}
}

func HandleClient(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Client disconnected or error:", err)
			return
		}

		msg := string(buffer[:n])
		fmt.Println("Received message:", msg)

		response := fmt.Sprintf("Bien reçu: %s", msg)
		conn.Write([]byte(response))
	}
}
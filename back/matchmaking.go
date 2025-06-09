package back

import (
	"encoding/json"
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
	var Pseudo string

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Client disconnected or error:", err)
			return
		}

		msg := string(buffer[:n])
		fmt.Println("Received message:", msg)

		var data map[string]string
		json.Unmarshal(buffer[:n], &data)

		switch data["type"] {
		case "join":
			Pseudo = data["pseudo"]
			addToQueue(Client{Conn:conn, Pseudo: Pseudo})

		case "move":
			fmt.Println(Pseudo, "a joué:", data["row"], data["col"])
		}
	}
}

func addToQueue(c Client) {
	mutex.Lock()
	queue = append(queue, c)
	fmt.Println("Joueur en file d'attente:", c.Pseudo)	

	if len(queue) >= 2 {
		p1 := queue[0]
		p2 := queue[1]
		queue = queue[2:]

		go startGame(p1, p2)
	}
	mutex.Unlock()
}

func startGame(p1, p2 Client) {
	fmt.Println(" Match trouvé :", p1.Pseudo, "vs", p2.Pseudo)
	
	matchInfo := map[string]string{
		"type": "start",
		"You": p1.Pseudo,
		"Opponent": p2.Pseudo,
		"symbole": "X",
		"turn" : "true",
	}

	data1, _:= json.Marshal(matchInfo)
	p1.Conn.Write(data1)

	matchInfo["You"] = p2.Pseudo
	matchInfo["Opponent"] = p1.Pseudo
	matchInfo["symbole"] = "O"
	matchInfo["turn"] = "false"
	data2, _ := json.Marshal(matchInfo)
	p2.Conn.Write(data2)
}
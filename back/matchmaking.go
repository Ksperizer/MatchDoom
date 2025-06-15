package back

import (
	"net/http"
	"time"
	"fmt"


	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {return true },
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	
}

func AddToWSQueue(client *WSClient) {
	hub.addToQueue(client)
}

func RemoveFromWSQueue(client *WSClient) {
	hub.removeFromQueue(client)
}

func GenerateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}

func GenerateGameID(p1, p2 string) string {
	return fmt.Sprintf("game_%s_vs_%s_%d", p1, p2, time.Now().UnixNano())
}



func init() {
	wsActiveGames = make(map[string]*WSGame)
}

func CheckWinner(board [3][3]string) string {
	// Lignes
	for i := 0; i < 3; i++ {
		if board[i][0] != "" && board[i][0] == board[i][1] && board[i][1] == board[i][2] {
			return board[i][0]
		}
	}
	
	// Colonnes
	for j := 0; j < 3; j++ {
		if board[0][j] != "" && board[0][j] == board[1][j] && board[1][j] == board[2][j] {
			return board[0][j]
		}
	}
	
	// Diagonales
	if board[0][0] != "" && board[0][0] == board[1][1] && board[1][1] == board[2][2] {
		return board[0][0]
	}
	if board[0][2] != "" && board[0][2] == board[1][1] && board[1][1] == board[2][0] {
		return board[0][2]
	}
	
	return ""
}

func IsBoardFull(board [3][3]string) bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if board[i][j] == "" {
				return false
			}
		}
	}
	return true
}





// func InitGameServer(){
// 	// execute python for launch game
// 	log.Println("Lancement du serveur Python pour le jeu...")

// 	if _, err := os.Stat("game/websocket.py"); os.IsNotExist(err) {
// 		log.Fatal("Fichier /game/websocket.py non trouvé")
// 	}

// 	wsCmd := exec.Command("python", "game/websocket.py")
// 	wsCmd.Stdout = os.Stdout
// 	wsCmd.Stderr = os.Stderr

// 	err := wsCmd.Start()
// 	if err != nil {
// 		log.Fatalf("Erreur lors du lancement du serveur Python: %v", err)

// 		wsCmd = exec.Command("python", "game/websocket.py")
// 		wsCmd.Stdout = os.Stdout
// 		wsCmd.Stderr = os.Stderr
// 		err = wsCmd.Start()
// 		if err != nil {
// 			log.Fatal("Impossible de lancer le serveur websocket.py:", err)
// 		}
// 	}

// 	log.Printf("websocket.py lancé avec PID %d", wsCmd.Process.Pid)

// 	time.Sleep(2 * time.Second) // cooldown python server ready 

// 	pythonWS := url.URL{Scheme: "ws", Host: "localhost:8081"}
// 	testConn, _, err := websocket.DefaultDialer.Dial(pythonWS.String(), nil)
// 	if err != nil {
// 		log.Printf(" Serveur Python WebSocket injoignable : %v", err)
// 	} else {
// 		testConn.Close()
// 		log.Println("Connexion au serveur OK")
// 	}
	
// }

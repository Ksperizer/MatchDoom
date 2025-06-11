package back

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"github.com/gorilla/websocket"
	"fmt"
	"time"
	"os/exec"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allows all origins (to be secured in production)
	},
}

type WSClient struct {
	Conn   *websocket.Conn
	Pseudo string
	ID     string
	Send   chan []byte
}

type WSGame struct {
	Player1 *WSClient
	Player2 *WSClient
	Board   [3][3]string
	Turn    string
}

var wsQueue []*WSClient
var wsActiveGames map[string]*WSGame
var wsMutex sync.Mutex
var wsGamesMutex sync.Mutex

func init() {
	wsActiveGames = make(map[string]*WSGame)
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &WSClient{
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	go client.WritePump()
	go client.ReadPump()
}

func (c *WSClient) ReadPump() {
	defer func() {
		c.Conn.Close()
		RemoveFromWSQueue(c)
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		var data map[string]string
		if err := json.Unmarshal(message, &data); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			continue
		}

		HandleWSMessage(c, data)
	}
}

func (c *WSClient) WritePump() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

func HandleWSMessage(client *WSClient, data map[string]string) {
	switch data["type"] {
	case "join":
		client.Pseudo = data["pseudo"]
		client.ID = GenerateWSClientID()
		AddToWSQueue(client)

	case "move":
		HandleWSMove(client, data)

	case "ping":
		SendWSMessage(client, map[string]string{"type": "pong"})
	}
}

func AddToWSQueue(client *WSClient) {
	wsMutex.Lock()
	defer wsMutex.Unlock()

	// check double pseudo
	for _, existing := range wsQueue {
		if existing.Pseudo == client.Pseudo {
			SendWSMessage(client, map[string]string{
				"type": "error",
				"message": "Pseudo déjà utilisé",
			})
			return
		}
	}

	wsQueue = append(wsQueue, client)
	log.Printf("Player %s added to WebSocket queue (%d total)", client.Pseudo, len(wsQueue))

	SendWSMessage(client, map[string]interface{}{
		"type": "queue",
		"message": "En attente d'un adversaire...",
		"position": len(wsQueue),
	})

	if len(wsQueue) >= 2 {
		p1 := wsQueue[0]
		p2 := wsQueue[1]
		wsQueue = wsQueue[2:]

		go StartWSGame(p1, p2)
	}
}

func RemoveFromWSQueue(client *WSClient) {
	wsMutex.Lock()
	defer wsMutex.Unlock()

	for i, c := range wsQueue {
		if c == client {
			wsQueue = append(wsQueue[:i], wsQueue[i+1:]...)
			log.Printf("Removed %s from queue", client.Pseudo)
			break
		}
	}
}

func StartWSGame(p1, p2 *WSClient) {
	gameID := GenerateGameID(p1.Pseudo, p2.Pseudo)

	game := &WSGame{
		Player1: p1,
		Player2: p2,
		Board:   [3][3]string{},
		Turn:    p1.Pseudo,
	}

	wsGamesMutex.Lock()
	wsActiveGames[gameID] = game
	wsGamesMutex.Unlock()

	log.Printf("WebSocket game started: %s vs %s", p1.Pseudo, p2.Pseudo)

	// Messages de début
	startMsg1 := map[string]interface{}{
		"type": "game_start",
		"game_id": gameID,
		"you": p1.Pseudo,
		"opponent": p2.Pseudo,
		"symbol": "X",
		"your_turn": true,
		"board": game.Board,
	}

	startMsg2 := map[string]interface{}{
		"type": "game_start",
		"game_id": gameID,
		"you": p2.Pseudo,
		"opponent": p1.Pseudo,
		"symbol": "O",
		"your_turn": false,
		"board": game.Board,
	}

	SendWSMessage(p1, startMsg1)
	SendWSMessage(p2, startMsg2)
}

func HandleWSMove(client *WSClient, data map[string]string) {
	gameID := data["game_id"]
	row := data["row"]
	col := data["col"]

	wsGamesMutex.Lock()
	game, exists := wsActiveGames[gameID]
	wsGamesMutex.Unlock()

	if !exists || game.Turn != client.Pseudo {
		SendWSMessage(client, map[string]string{
			"type": "error",
			"message": "Coup invalide",
		})
		return
	}

	// is move valid
	rowInt := ParseCoord(row)
	colInt := ParseCoord(col)
	
	if rowInt < 0 || rowInt > 2 || colInt < 0 || colInt > 2 || game.Board[rowInt][colInt] != "" {
		SendWSMessage(client, map[string]string{
			"type": "error",
			"message": "Position invalide",
		})
		return
	}

	// play the move
	symbol := "X"
	if client == game.Player2 {
		symbol = "O"
	}
	game.Board[rowInt][colInt] = symbol

	// victory check
	winner := CheckWinner(game.Board)
	gameOver := winner != "" || IsBoardFull(game.Board)

	// Change turn 
	if !gameOver {
		if game.Turn == game.Player1.Pseudo {
			game.Turn = game.Player2.Pseudo
		} else {
			game.Turn = game.Player1.Pseudo
		}
	}

	
	moveMsg := map[string]interface{}{
		"type": "move_played",
		"player": client.Pseudo,
		"row": row,
		"col": col,
		"symbol": symbol,
		"board": game.Board,
		"next_turn": game.Turn,
		"game_over": gameOver,
		"winner": winner,
	}

	SendWSMessage(game.Player1, moveMsg)
	SendWSMessage(game.Player2, moveMsg)

	if gameOver {
		wsGamesMutex.Lock()
		delete(wsActiveGames, gameID)
		wsGamesMutex.Unlock()
		log.Printf("Game %s finished. Winner: %s", gameID, winner)
	}
}

func SendWSMessage(client *WSClient, data interface{}) {
	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return
	}

	select {
	case client.Send <- message:
	default:
		close(client.Send)
	}
}

func GenerateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}


func GenerateWSClientID() string {
	return GenerateClientID() // Reuse existing function
}

func GenerateGameID(p1, p2 string) string {
	return p1 + "_vs_" + p2
}

func ParseCoord(coord string) int {
	switch coord {
	case "0": return 0
	case "1": return 1
	case "2": return 2
	default: return -1
	}
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

func InitGameServer(){
	// execute python for launch game
	launchCmd := exec.Command("python", "game/launch.py")
	err1 := launchCmd.Start()
	if err1 != nil {
		log.Printf("Erreur de lancement du serveur Python: %v", err1)
	}else {
		log.Printf("launch.py lancer avec PID %d", launchCmd.Process.Pid)
	}

	// execute python for launch websocket
	wsCmd := exec.Command("python", "game/websocket.py")
	err2 := wsCmd.Start()
	if err2 != nil {
		log.Printf("Erreur de lancement du serveur WebSocket: %v", err2)
	} else {
		log.Printf("websocket.py lancer avec PID %d", wsCmd.Process.Pid)
	}
}
package back

import (
	"MatchDoom/data"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	Conn       *websocket.Conn
	Pseudo     string
	UserID     string
	ID         string
	Send       chan []byte
	GameID     string
	Connected  time.Time
	LastPing   time.Time
	PythonConn *websocket.Conn
	IsActive   bool
	Rating     int
	mutex      sync.RWMutex
}

type WSGame struct {
	ID         string
	Player1    *WSClient
	Player2    *WSClient
	Board      [3][3]string
	Turn       string
	Created    time.Time
	MatchID    uint
	IsFinished bool
	LastMove   *GameMove
}

type GameMove struct {
	Player string
	Row    int
	Col    int
	Symbol string
	Time   time.Time
}

type ConnectionStats struct {
	TotalConnections   int
	ActiveConnections  int
	ActiveGames        int
	QueueSize          int
	PythonServerStatus string
	Uptime             time.Duration
	LastCleanup        time.Time
}

type MatchmakingHub struct {
	// active connection
	clients map[string]*WSClient

	queue []*WSClient

	activeGames map[string]*WSGame

	//channels for communication
	register   chan *WSClient
	unregister chan *WSClient
	broadcast  chan []byte

	// mutex for concurrent access
	clientMutex sync.RWMutex
	queueMutex  sync.RWMutex
	gamesMutex  sync.RWMutex

	// stats
	stats     ConnectionStats
	startTime time.Time

	//config
	maxConnections  int
	queueTimeout    time.Duration
	gameTimeout     time.Duration
	cleanupInterval time.Duration
}

var hub *MatchmakingHub

var wsQueue []*WSClient
var wsActiveGames map[string]*WSGame
var wsMutex sync.Mutex
var wsGamesMutex sync.Mutex

func Init() {
	wsActiveGames = make(map[string]*WSGame)

	hub = &MatchmakingHub{
		clients:         make(map[string]*WSClient),
		queue:           make([]*WSClient, 0),
		activeGames:     make(map[string]*WSGame),
		register:        make(chan *WSClient, 256),
		unregister:      make(chan *WSClient, 256),
		broadcast:       make(chan []byte, 256),
		maxConnections:  1000,
		queueTimeout:    5 * time.Minute,
		gameTimeout:     30 * time.Minute,
		cleanupInterval: 2 * time.Minute,
		startTime:       time.Now(),
	}

	go hub.Run()
	go hub.StartMaintenanceWorker()
}

func (h *MatchmakingHub) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Println(" Hub de matchmaking lancé...")

	for {
		select {
		case client := <-h.register:
			h.handleClientRegister(client)

		case client := <-h.unregister:
			h.handleClientUnregister(client)

		case message := <-h.broadcast:
			h.handleBroadcast(message)

		case <-ticker.C:
			h.updateStats()
			h.pingClients()
		}
	}
}

// Gestionnary connexion
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf(" Erreur upgrade WebSocket: %v", err)
		return
	}

	// check max connections
	hub.clientMutex.RLock()
	activeCount := len(hub.clients)
	hub.clientMutex.RUnlock()

	if activeCount >= hub.maxConnections {
		log.Printf("Limite de connexions atteinte (%d)", hub.maxConnections)
		clientConn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "Serveur temporairement surchargé",
		})
		clientConn.Close()
		return
	}

	// create new client
	client := &WSClient{
		Conn:      clientConn,
		ID:        GenerateClientID(),
		Send:      make(chan []byte, 256),
		Connected: time.Now(),
		LastPing:  time.Now(),
		IsActive:  true,
		Rating:    1200,
	}

	log.Printf("Nouvelle connexion WebSocket: %s", client.ID)

	// connect to Python server
	if !hub.connectToPythonServer(client) {
		clientConn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "Serveur de jeu temporairement indisponible",
		})
		clientConn.Close()
		return
	}

	// Save client in hub
	hub.register <- client

	//Launch client goroutines
	go client.writePump()
	go client.readPump()
	go client.proxyToPython()
	go client.proxyFromPython()
}

func (h *MatchmakingHub) connectToPythonServer(client *WSClient) bool {
	pythonURL := url.URL{Scheme: "ws", Host: "localhost:8081"}

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		pythonConn, _, err := websocket.DefaultDialer.Dial(pythonURL.String(), nil)
		if err != nil {
			log.Printf("Tentative %d de connexion au serveur Python échouée: %v", i+1, err)
			if i == maxRetries-1 {
				return false
			}
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		client.PythonConn = pythonConn
		log.Printf("Connexion python établie pour le client %s", client.ID)
		return true
	}

	return false
}

func (h *MatchmakingHub) handleClientRegister(client *WSClient) {
	h.clientMutex.Lock()
	h.clients[client.ID] = client
	h.clientMutex.Unlock()

	log.Printf(" client enregistré: %s (total: %d)", client.ID, len(h.clients))

	welcome := map[string]interface{}{
		"type":      "connected",
		"message":   "Connexion réussie à matchdoom",
		"client_id": client.ID,
		"server_info": map[string]interface{}{
			"version":  "1.0.0",
			"features": []string{"matchmaking", "ratings", "persistence", "reconnection"},
		},
	}
	client.SendMessage(welcome)
}

func (h *MatchmakingHub) handleClientUnregister(client *WSClient) {
	h.clientMutex.Lock()
	delete(h.clients, client.ID)
	h.clientMutex.Unlock()

	// Remove from queue
	h.removeFromQueue(client)

	h.handleGameDisconnection(client)

	if client.PythonConn != nil {
		client.PythonConn.Close()
	}
	close(client.Send)

	log.Printf("Client deconnected: %s (%s)", client.ID, client.Pseudo)
}

func (h *MatchmakingHub) handleBroadcast(message []byte) {
	h.clientMutex.RLock()
	defer h.clientMutex.RUnlock()

	for _, client := range hub.clients {
		if client.IsActive {
			select {
			case client.Send <- message:
			default:
				client.IsActive = false
			}
		}
	}
}

func (c *WSClient) readPump() {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.mutex.Lock()
		c.LastPing = time.Now()
		c.mutex.Unlock()
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Erreur de lecture WebSocket: %v", err)
			}
			break
		}
		c.handleMessage(message)
	}
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(60 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Erreur d'écriture WebSocket: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *WSClient) proxyToPython() {
	for {
		if c.PythonConn == nil {
			break
		}

		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		// Intercepter certains messages pour traitement côté Go
		if c.interceptMessage(message) {
			continue
		}

		// Transférer vers Python
		if err := c.PythonConn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf(" Erreur proxy vers Python: %v", err)
			break
		}
	}
}

func (c *WSClient) proxyFromPython() {
	for {
		if c.PythonConn == nil {
			break
		}

		_, message, err := c.PythonConn.ReadMessage()
		if err != nil {
			break
		}

		// Intercepter certains messages pour traitement côté Go
		if c.processFromPython(message) {
			continue
		}

		// Transférer vers le client
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Erreur proxy vers client: %v", err)
			break
		}
	}
}

func (c *WSClient) handleMessage(message []byte) {
	var msgData map[string]interface{}
	if err := json.Unmarshal(message, &msgData); err != nil {
		log.Printf("Erreur JSON: %v", err)
		return
	}

	msgType, ok := msgData["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "join":
		c.handleJoin(msgData)
	case "ping":
		c.handlePing()
	case "get_stats":
		c.handleGetStats()
	default:
		// Transfer message to Python server
		if c.PythonConn != nil {
			c.PythonConn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func (c *WSClient) interceptMessage(message []byte) bool {
	var msgData map[string]interface{}
	if err := json.Unmarshal(message, &msgData); err != nil {
		return false
	}

	msgType, ok := msgData["type"].(string)
	if !ok {
		return false
	}

	switch msgType {
	case "join":
		return c.handleJoin(msgData)
	case "ping":
		return c.handlePing()
	case "get_stats":
		return c.handleGetStats()
	}

	return false
}

func (c *WSClient) processFromPython(message []byte) bool {
	var msgData map[string]interface{}
	if err := json.Unmarshal(message, &msgData); err != nil {
		return false
	}

	msgType, ok := msgData["type"].(string)
	if !ok {
		return false
	}

	switch msgType {
	case "game_start":
		c.handleGameStartFromPython(msgData)
	case "move_played":
		c.handleMoveFromPython(msgData)
	case "game_end":
		c.handleGameEndFromPython(msgData)
	}

	return false // Laisser passer le message
}

func (c *WSClient) handleJoin(msgData map[string]interface{}) bool {
	pseudo, ok := msgData["pseudo"].(string)
	if !ok {
		c.SendError("Pseudo requis")
		return true
	}

	// check if the client is already connected
	user, err := data.GetUserByPseudo(pseudo)
	if err != nil {
		c.SendError("Utilisateur non trouvé en base de données")
		return true
	}

	// update client information
	c.mutex.Lock()
	c.Pseudo = pseudo
	c.UserID = fmt.Sprintf("%d", user.ID)
	if user.TotalGames > 0 {
		winRate := float64(user.Wins) / float64(user.TotalGames)
		c.Rating = int(1200 + (winRate-0.5)*800)
	}
	c.mutex.Unlock()

	// update last_seen
	data.UpdateUserLastSeen(user.ID)
	log.Printf("Joueur connecté: %s (ID: %d, Rating: %d)", pseudo, user.ID, c.Rating)

	// add client to matchmaking queue
	hub.addToQueue(c)

	return false
}

func (c *WSClient) handlePing() bool {
	c.mutex.Lock()
	c.LastPing = time.Now()
	c.mutex.Unlock()

	c.SendMessage(map[string]string{"type": "pong"})
	return true
}

func (c *WSClient) handleGetStats() bool {
	stats := hub.GetCurrentStats()
	c.SendMessage(map[string]interface{}{
		"type": "stats",
		"data": stats,
	})
	return true
}

func (c *WSClient) handleGameStartFromPython(msgData map[string]interface{}) {
	gameID, _ := msgData["game_id"].(string)
	opponent, _ := msgData["opponent"].(string)

	c.mutex.Lock()
	c.GameID = gameID
	c.mutex.Unlock()

	log.Printf("Partie commencée: %s vs %s", c.Pseudo, opponent)
}

func (c *WSClient) handleMoveFromPython(msgData map[string]interface{}) {
	gameOver, _ := msgData["game_over"].(bool)
	if gameOver {
		winner, _ := msgData["winner"].(string)
		log.Printf("Partie terminée: %s, vainqueur: %s", c.GameID, winner)
	}
}

func (c *WSClient) handleGameEndFromPython(msgData map[string]interface{}) {
	c.mutex.Lock()
	c.GameID = ""
	c.mutex.Unlock()

	winner, _ := msgData["winner"].(string)
	log.Printf("Fin de partie pour %s, résultat: %s", c.Pseudo, winner)
}

func (c *WSClient) SendMessage(data interface{}) {
	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("Erreur marshaling JSON: %v", err)
		return
	}

	select {
	case c.Send <- message:
	default:
		// channel full, ping inactive client
		c.IsActive = false
		log.Printf("Client %s non réactif", c.ID)
	}
}

func (c *WSClient) SendError(message string) {
	c.SendMessage(map[string]string{
		"type":    "error",
		"message": message,
	})
}

// === GESTION DE LA QUEUE ===

func (h *MatchmakingHub) addToQueue(client *WSClient) {
	h.queueMutex.Lock()
	defer h.queueMutex.Unlock()

	// Chech twice client
	for _, existing := range h.queue {
		if existing.Pseudo == client.Pseudo {
			client.SendError("Déjà en file d'attente")
			return
		}
	}

	h.queue = append(h.queue, client)

	// Sync with global variables compatibility
	wsMutex.Lock()
	wsQueue = h.queue
	wsMutex.Unlock()

	position := len(h.queue)
	log.Printf("%s en queue (pos: %d, rating: %d)", client.Pseudo, position, client.Rating)

	client.SendMessage(map[string]interface{}{
		"type":           "queue_joined",
		"message":        "En attente d'un adversaire...",
		"position":       position,
		"estimated_wait": h.estimateWaitTime(position),
		"your_rating":    client.Rating,
	})
}

func (h *MatchmakingHub) removeFromQueue(client *WSClient) {
	h.queueMutex.Lock()
	defer h.queueMutex.Unlock()

	for i, c := range h.queue {
		if c == client {
			h.queue = append(h.queue[:i], h.queue[i+1:]...)

			// Sync with global variables
			wsMutex.Lock()
			wsQueue = h.queue
			wsMutex.Unlock()

			log.Printf("%s retiré de la queue", client.Pseudo)
			break
		}
	}
}

func (h *MatchmakingHub) estimateWaitTime(position int) string {
	if position <= 1 {
		return "< 30s"
	} else if position <= 3 {
		return "30s - 1min"
	} else if position <= 6 {
		return "1-2min"
	}
	return "2-5min"
}

// === Gestionnary of Disconnect  ===

func (h *MatchmakingHub) handleGameDisconnection(client *WSClient) {
	if client.GameID == "" {
		return
	}

	//notify Python server
	if client.PythonConn != nil {
		disconnectMsg := map[string]interface{}{
			"type":    "player_disconnect",
			"pseudo":  client.Pseudo,
			"game_id": client.GameID,
		}
		msgBytes, _ := json.Marshal(disconnectMsg)
		client.PythonConn.WriteMessage(websocket.TextMessage, msgBytes)
	}

	log.Printf("🔌 Déconnexion en jeu: %s (partie: %s)", client.Pseudo, client.GameID)
}

//MAINTENANCE & CLEANUP

func (h *MatchmakingHub) StartMaintenanceWorker() {
	ticker := time.NewTicker(h.cleanupInterval)
	defer ticker.Stop()

	log.Println("Worker de maintenance démarré")

	for {
		select {
		case <-ticker.C:
			h.performMaintenance()
		}
	}
}

func (h *MatchmakingHub) performMaintenance() {
	now := time.Now()

	// clean inactive clients
	h.cleanupInactiveClients(now)

	// clean queue
	h.cleanupQueue(now)

	// clean old database entries
	h.cleanupDatabase()

	// update stats
	h.updateStats()

	h.stats.LastCleanup = now

	log.Printf("Maintenance effectuée - Clients: %d, Queue: %d, Uptime: %v",
		len(h.clients), len(h.queue), time.Since(h.startTime).Truncate(time.Second))
}

func (h *MatchmakingHub) cleanupInactiveClients(now time.Time) {
	h.clientMutex.Lock()
	defer h.clientMutex.Unlock()

	for clientID, client := range h.clients {
		client.mutex.RLock()
		timeSinceLastPing := now.Sub(client.LastPing)
		client.mutex.RUnlock()

		if timeSinceLastPing > 5*time.Minute || !client.IsActive {
			log.Printf("Nettoyage client inactif: %s", clientID)
			delete(h.clients, clientID)

			if client.PythonConn != nil {
				client.PythonConn.Close()
			}
			close(client.Send)
		}
	}
}

func (h *MatchmakingHub) cleanupQueue(now time.Time) {
	h.queueMutex.Lock()
	defer h.queueMutex.Unlock()

	originalLen := len(h.queue)
	newQueue := make([]*WSClient, 0, originalLen)

	for _, client := range h.queue {
		client.mutex.RLock()
		timeInQueue := now.Sub(client.Connected)
		client.mutex.RUnlock()

		if timeInQueue < h.queueTimeout && client.IsActive {
			newQueue = append(newQueue, client)
		} else {
			log.Printf("Timeout queue: %s", client.Pseudo)
			client.SendError("Timeout de la file d'attente")
		}
	}

	h.queue = newQueue

	// synchro with global variables
	wsMutex.Lock()
	wsQueue = h.queue
	wsMutex.Unlock()

	if len(h.queue) != originalLen {
		log.Printf("Queue nettoyée: %d -> %d", originalLen, len(h.queue))
	}
}

func (h *MatchmakingHub) cleanupDatabase() {
	// Clean old queue entries from the database
	if err := data.CleanOldQueue(); err != nil {
		log.Printf("Erreur nettoyage base queue: %v", err)
	}
}

func (h *MatchmakingHub) updateStats() {
	h.clientMutex.RLock()
	activeConnections := len(h.clients)
	h.clientMutex.RUnlock()

	h.queueMutex.RLock()
	queueSize := len(h.queue)
	h.queueMutex.RUnlock()

	h.gamesMutex.RLock()
	activeGames := len(h.activeGames)
	h.gamesMutex.RUnlock()

	// check Python server status
	pythonStatus := h.checkPythonServerStatus()

	h.stats = ConnectionStats{
		TotalConnections:   h.stats.TotalConnections, 
		ActiveConnections:  activeConnections,
		ActiveGames:        activeGames,
		QueueSize:          queueSize,
		PythonServerStatus: pythonStatus,
		Uptime:             time.Since(h.startTime),
		LastCleanup:        h.stats.LastCleanup,
	}
}

func (h *MatchmakingHub) checkPythonServerStatus() string {
	pythonURL := url.URL{Scheme: "ws", Host: "localhost:8081"}
	conn, _, err := websocket.DefaultDialer.Dial(pythonURL.String(), nil)
	if err != nil {
		return "offline"
	}
	conn.Close()
	return "online"
}

func (h *MatchmakingHub) pingClients() {
	h.clientMutex.RLock()
	defer h.clientMutex.RUnlock()

	pingMsg := map[string]string{"type": "server_ping"}
	pingBytes, _ := json.Marshal(pingMsg)

	for _, client := range h.clients {
		if client.IsActive {
			select {
			case client.Send <- pingBytes:
			default:
				client.IsActive = false
			}
		}
	}
}

// Public Api

func GetProxyStats() map[string]interface{} {
	if hub == nil {
		return map[string]interface{}{
			"status": "initializing",
		}
	}

	stats := hub.GetCurrentStats()

	return map[string]interface{}{
		"active_connections":   stats.ActiveConnections,
		"queue_size":           stats.QueueSize,
		"active_games":         stats.ActiveGames,
		"python_server_status": stats.PythonServerStatus,
		"uptime_seconds":       int(stats.Uptime.Seconds()),
		"last_cleanup":         stats.LastCleanup,
		"proxy_type":           "enhanced_go_python",
		"status":               "healthy",
		"max_connections":      hub.maxConnections,
	}
}

func (h *MatchmakingHub) GetCurrentStats() ConnectionStats {
	return h.stats
}

// === FONCTIONS DE COMPATIBILITÉ ===


func StartWSGame(p1, p2 *WSClient) {
	// in python
	log.Printf("Démarrage de partie géré par Python: %s vs %s", p1.Pseudo, p2.Pseudo)
}

func HandleWSMessage(client *WSClient, msgData map[string]string) {
	// convert new format
	dataInterface := make(map[string]interface{})
	for k, v := range msgData {
		dataInterface[k] = v
	}
	client.handleJoin(dataInterface)
}

func HandleWSMove(client *WSClient, data map[string]string) {
	// move in python file
	log.Printf("🎯 Mouvement transféré vers Python: %s", client.Pseudo)
}

func SendWSMessage(client *WSClient, data interface{}) {
	client.SendMessage(data)
}

func GenerateWSClientID() string {
	return GenerateClientID()
}

func ParseCoord(coord string) int {
	switch coord {
	case "0":
		return 0
	case "1":
		return 1
	case "2":
		return 2
	default:
		return -1
	}
}

// === GESTION DU SERVEUR PYTHON ===

func InitGameServer() {
	log.Println("Vérification du serveur Python WebSocket...")

	// check if the websocket.py file exists
	if _, err := os.Stat("game/websocket.py"); os.IsNotExist(err) {
		log.Fatal("Fichier game/websocket.py non trouvé")
	}

	// check if the Python server is already running
	if isPythonServerRunning() {
		log.Println("Serveur Python déjà en cours")
		return
	}

	log.Println("Démarrage du serveur Python WebSocket...")

	//start the Python server
	cmd := exec.Command("python", "game/websocket.py")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Printf("Erreur démarrage serveur Python: %v", err)
		// Retry with python3 if python fails
		cmd = exec.Command("python3", "game/websocket.py")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			log.Fatal("Impossible de lancer le serveur Python:", err)
		}
	}

	log.Printf(" Serveur Python lancé avec PID %d", cmd.Process.Pid)

	// wait for the server to be ready
	maxWait := 30 // 30 secondes max
	for i := 0; i < maxWait; i++ {
		time.Sleep(1 * time.Second)
		if isPythonServerRunning() {
			log.Println(" Serveur Python prêt et accessible")
			return
		}
		if i%5 == 0 && i > 0 {
			log.Printf("Attente du serveur Python... (%d/%ds)", i, maxWait)
		}
	}

	log.Fatal(" Timeout: Serveur Python non accessible après 30 secondes")
}

func isPythonServerRunning() bool {
	pythonURL := url.URL{Scheme: "ws", Host: "localhost:8081"}
	conn, _, err := websocket.DefaultDialer.Dial(pythonURL.String(), nil)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

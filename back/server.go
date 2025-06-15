package back

import (
	"MatchDoom/data"
	"MatchDoom/handlers"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func Server() error {
	InitGameServer()


	r := mux.NewRouter()

	// Redirection vers /accueil si chemin "/"
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/accueil", http.StatusMovedPermanently)
	}).Methods("GET")

	setupStaticRoutes(r)
	setupPageRoutes(r)
	setupAPIRoutes(r)

	// WebSocket Game 
	r.HandleFunc("/game/ws", HandleWebSocket)

	port := "8080"
	log.Printf("Server running: http://localhost:%s", port)
	log.Printf("WebSocket Python Game: ws://localhost:8081/game/ws")
	log.Println("Ready!")

	return http.ListenAndServe(":"+port, r)
}

func setupStaticRoutes(r *mux.Router) {
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("template/css/"))))
	r.PathPrefix("/image/").Handler(http.StripPrefix("/image/", http.FileServer(http.Dir("template/ressource/image/"))))
	r.PathPrefix("/script/").Handler(http.StripPrefix("/script/", http.FileServer(http.Dir("template/script/"))))
}

func setupPageRoutes(r *mux.Router) {
	r.HandleFunc("/accueil", AccueilHandle).Methods("GET")
	r.HandleFunc("/connexion", ConnexionHandle).Methods("GET")
	r.HandleFunc("/profil", ProfilHandle).Methods("GET")
	
	//r.HandleFunc("/game", GameHandle).Methods("GET")
}

func setupAPIRoutes(r *mux.Router) {
	api := r.PathPrefix("/api").Subrouter()

	// Auth
	api.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
	api.HandleFunc("/login", handlers.LoginUser).Methods("POST")

	// User Profile
	api.HandleFunc("/profile", handlers.GetProfile).Methods("GET")

	// Stats
	api.HandleFunc("/update-stats", handlers.UpdateStats).Methods("POST")
	api.HandleFunc("/leaderboard", handlers.GetLeaderboard).Methods("GET")
	api.HandleFunc("/stats", handlers.GetStats).Methods("GET")

	// Queue and Matchmaking
	api.HandleFunc("/queue/join", handlers.JoinQueue).Methods("POST")
	api.HandleFunc("/matches/active", handlers.GetActiveMatches).Methods("GET")

	// WebSocket
	api.HandleFunc("/ws", HandleWebGameWS).Methods("GET")

	// PROXY WebSocket to Python Game
	api.HandleFunc("/proxy/stats", getProxyStats).Methods("GET")
	api.HandleFunc("/python/status", getPythonServerStatus).Methods("GET")



	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		stats, err := data.GetGameStats()
		if err != nil {
			http.Error(w, "Erreur base de données", http.StatusInternalServerError)
			return
		}

		proxyStats := GetProxyStats()

		response := map[string]interface{}{
			"status":            "ok",
			"database":          "connected",
			"python_game":       "ws://localhost:8081/game/ws",
			"python_game_status": proxyStats["status"],
			"total_users":       stats["total_users"],
			"active_matches":    stats["active_matches"],
			"total_matches":     stats["total_matches"],
			"queue_count":       stats["queue_count"],
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")
}

// front 

func AccueilHandle(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("template/html/accueil.html")
	if err != nil {
		log.Printf("Error loading accueil.html: %v", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}

	// get stats from the database
	stats, err := data.GetGameStats()
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		stats = map[string]int{
			"total_users":    0,
			"active_matches": 0,
			"total_matches":  0,
		}
	}

	wsURL := os.Getenv("PY_WS_URL")
	if wsURL == "" {
		wsURL = "ws://localhost:8081"
	}

	templateData := struct {
		Title         string
		TotalPlayers  int
		ActiveMatches int
		TotalMatches  int
		PythonGameURL string
	}{
		Title:         "MatchDoom - Tic Tac Toe Online",
		TotalPlayers:  stats["total_users"],
		ActiveMatches: stats["active_matches"],
		TotalMatches:  stats["total_matches"],
		PythonGameURL: "Pour jouer: connectez-vous à " + wsURL,
	}

	tmpl.Execute(w, templateData)
}

func ConnexionHandle(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("template/html/connexion.html")
	if err != nil {
		log.Printf("Error loading connexion.html: %v", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func ProfilHandle(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("template/html/profil.html")
	if err != nil {
		log.Printf("Error loading profil.html: %v", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func HandleWebGameWS(w http.ResponseWriter, r *http.Request) {
	HandleWebSocket(w, r)
}

// stats 
type GameStats struct {
	TotalPlayers  int
	ActiveMatches int
	TotalMatches  int
}

func getProxyStats(w http.ResponseWriter, r *http.Request) {
	stats := GetProxyStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func getPythonServerStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"python_server" : "ws://localhost:8081",
		"status" :"checking",
	}

	pythonURL := url.URL{Scheme: "ws", Host: "localhost:8081"}
	conn, _, err := websocket.DefaultDialer.Dial(pythonURL.String(), nil)
	if err != nil {
		status["status"] = "offline"
		status["error"] = err.Error()
	}else {
		conn.Close()
		status["status"] = "online"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}



// === ARRÊT ===

func Shutdown() {
	log.Println(" Arrêt du hub de matchmaking...")
	
	if hub != nil {
		// close all active connections
		hub.clientMutex.Lock()
		for _, client := range hub.clients {
			if client.PythonConn != nil {
				client.PythonConn.Close()
			}
			close(client.Send)
		}
		hub.clientMutex.Unlock()
	}
	
	log.Println(" Hub fermé proprement")
}





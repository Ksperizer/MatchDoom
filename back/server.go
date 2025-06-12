package back

import (
	"MatchDoom/data"
	"MatchDoom/handlers"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
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

	// WebSocket Game (client Python ou JS)
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
	r.HandleFunc("/api/register", handlers.RegisterUser).Methods("POST")
	r.HandleFunc("/api/login", handlers.LoginUser).Methods("POST")
	//r.HandleFunc("/game", GameHandle).Methods("GET")
}

func setupAPIRoutes(r *mux.Router) {
	api := r.PathPrefix("/api").Subrouter()

	// Authe
	api.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
	api.HandleFunc("/login", handlers.LoginUser).Methods("POST")
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

	// Health check amélioré
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		stats, err := data.GetGameStats()
		if err != nil {
			http.Error(w, "Erreur base de données", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"status":            "ok",
			"database":          "connected",
			"python_game":       "ws://localhost:8081/game/ws",
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

// ===== Pages HTML =====

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

// ===== Statistiques =====
type GameStats struct {
	TotalPlayers  int
	ActiveMatches int
	TotalMatches  int
}

// func getGameStats() GameStats {
// 	var stats GameStats

// 	err := data.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalPlayers)
// 	if err != nil {
// 		log.Println("Erreur getGameStats (users):", err)
// 	}

// 	err = data.DB.QueryRow("SELECT COUNT(*) FROM matches WHERE is_finished = FALSE").Scan(&stats.ActiveMatches)
// 	if err != nil {
// 		log.Println("Erreur getGameStats (active matches):", err)
// 	}

// 	err = data.DB.QueryRow("SELECT COUNT(*) FROM matches").Scan(&stats.TotalMatches)
// 	if err != nil {
// 		log.Println("Erreur getGameStats (total matches):", err)
// 	}

// 	return stats
// }

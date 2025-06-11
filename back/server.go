package back

import (
	"MatchDoom/data"
	"MatchDoom/handlers"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func Server() error {
	InitGameServer()
	data.InitDB()

	r := mux.NewRouter()

	// Redirection vers /accueil si chemin "/"
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/accueil", http.StatusMovedPermanently)
	}).Methods("GET")

	setupStaticRoutes(r)
	setupPageRoutes(r)
	setupAPIRoutes(r)

	// WebSocket Game (client Python ou JS)
	//r.HandleFunc("/game/ws", HandleGameWebSocket)

	port := "8080"
	log.Printf("Server running: http://localhost:%s", port)
	log.Printf("WebSocket Python Game: ws://localhost:%s/game/ws", port)
	log.Printf("Web interface: http://localhost:%s/accueil", port)
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
}

func setupAPIRoutes(r *mux.Router) {
	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
	api.HandleFunc("/login", handlers.LoginUser).Methods("POST")
	api.HandleFunc("/update-stats", handlers.UpdateStats).Methods("POST")
	api.HandleFunc("/ws", HandleWebGameWS).Methods("GET")

	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","python_game":"ws://localhost:8080/game/ws"}`))
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

	stats := getGameStats()
	data := struct {
		Title         string
		TotalPlayers  int
		ActiveMatches int
		TotalMatches  int
		PythonGameURL string
	}{
		Title:         "MatchDoom - Tic Tac Toe Online",
		TotalPlayers:  stats.TotalPlayers,
		ActiveMatches: stats.ActiveMatches,
		TotalMatches:  stats.TotalMatches,
		PythonGameURL: "Pour jouer: lancez le client Python et connectez-vous à ws://localhost:8080/game/ws",
	}

	tmpl.Execute(w, data)
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

// ===== WebSocket =====
func HandleWebGameWS(w http.ResponseWriter, r *http.Request) {
	HandleWebSocket(w, r)
}

// ===== Statistiques =====
type GameStats struct {
	TotalPlayers  int
	ActiveMatches int
	TotalMatches  int
}

func getGameStats() GameStats {
	var stats GameStats
	data.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalPlayers)
	data.DB.QueryRow("SELECT COUNT(*) FROM matches WHERE is_finished = FALSE").Scan(&stats.ActiveMatches)
	data.DB.QueryRow("SELECT COUNT(*) FROM matches").Scan(&stats.TotalMatches)
	return stats
}

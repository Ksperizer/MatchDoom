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

	r := mux.NewRouter()

	
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/accueil", http.StatusMovedPermanently)
	}).Methods("GET")

	
	setupStaticRoutes(r)

	
	setupMainRoutes(r)

	// API REST pour auth et stats
	setupAPIRoutes(r)

	
	r.HandleFunc("/game/ws", HandleGameWebSocket)

	port := "8080"
	log.Printf("Server running: http://localhost:%s", port)
	log.Printf("Python game WebSocket: ws://localhost:%s/game/ws", port)
	log.Printf("Web interface: http://localhost:%s/accueil", port)
	log.Println(" Ready!")

	return http.ListenAndServe(":"+port, r)
}

func setupStaticRoutes(r *mux.Router) {
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("template/css/"))))
	r.PathPrefix("/image/").Handler(http.StripPrefix("/image/", http.FileServer(http.Dir("template/ressource/image/"))))
	r.PathPrefix("/script/").Handler(http.StripPrefix("/script/", http.FileServer(http.Dir("template/script/"))))
}

func setupMainRoutes(r *mux.Router) {
	r.HandleFunc("/accueil", AccueilHandle).Methods("GET")
	r.HandleFunc("/connexion", ConnexionHandle).Methods("GET")
	r.HandleFunc("/profil", ProfilHandle).Methods("GET")
	
	// Auth handlers (formulaires web)
	r.HandleFunc("/login", loginUser).Methods("POST")
	r.HandleFunc("/register", addUser).Methods("POST")
}

func setupAPIRoutes(r *mux.Router) {
	api := r.PathPrefix("/api").Subrouter()

	// API REST pour l'interface web
	api.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
	api.HandleFunc("/login", handlers.LoginUser).Methods("POST")
	api.HandleFunc("/update-stats", handlers.UpdateStats).Methods("POST")
	
	
	api.HandleFunc("/ws", HandleWebGameWS)
	
	// Health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "python_game": "ws://localhost:8080/game/ws"}`))
	}).Methods("GET")
}

// Handlers pour les pages HTML
func AccueilHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/accueil" {
		http.NotFound(w, r)
		return
	}
	
	tmpl, err := template.ParseFiles("template/html/accueil.html")
	if err != nil {
		log.Printf("Error parsing template %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	
	stats := getGameStats()
	
	data := struct {
		Title           string
		TotalPlayers    int
		ActiveMatches   int
		TotalMatches    int
		PythonGameURL   string
	}{
		Title:         "MatchDoom - Tic Tac Toe Online",
		TotalPlayers:  stats.TotalPlayers,
		ActiveMatches: stats.ActiveMatches,
		TotalMatches:  stats.TotalMatches,
		PythonGameURL: "Pour jouer: lancez le client Python et connectez-vous à ws://localhost:8080/game/ws",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

func ConnexionHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/connexion" {
		http.NotFound(w, r)
		return
	}
	
	tmpl, err := template.ParseFiles("template/html/connexion.html")
	if err != nil {
		log.Printf("Error parsing template %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

func ProfilHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/profil" {
		http.NotFound(w, r)
		return
	}
	
	tmpl, err := template.ParseFiles("template/html/profil.html")
	if err != nil {
		log.Printf("Error parsing template %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

// WebSocket for web game
func HandleWebGameWS(w http.ResponseWriter, r *http.Request) {
	
	HandleGameWebSocket(w, r)
}

// Handlers temporaires pour l'auth
func loginUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Login attempt via form")
	http.Redirect(w, r, "/accueil", http.StatusSeeOther)
}

func addUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Registration attempt via form")
	http.Redirect(w, r, "/connexion", http.StatusSeeOther)
}


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
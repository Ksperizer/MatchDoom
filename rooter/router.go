package rooter

import (
	"MatchDoom/handlers"
	"MatchDoom/back"
	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	r := mux.NewRouter()

	// API REST auth and stats 
	r.HandleFunc("/api/register", handlers.RegisterUser).Methods("POST")
	r.HandleFunc("/api/login", handlers.LoginUser).Methods("POST")
	r.HandleFunc("/api/update-stats", handlers.UpdateStats).Methods("POST")

	// WebSocket for the game
	r.HandleFunc("/api/ws", back.HandleWebSocket)

	return r
}
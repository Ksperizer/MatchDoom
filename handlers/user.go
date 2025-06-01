package handlers

import (
	"MatchDoom/data"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Pseudo   string `json:"pseudo"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	json.NewDecoder(r.Body).Decode(&req)

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Erreur lors du hash", 500) // Internal Server Error
		return
	}

	_, err = data.DB.Exec("INSERT INTO users (pseudo, password_hash, email) VALUES (?, ?, ?)", req.Pseudo, hash, req.Email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur SQL : %v", err), http.StatusInternalServerError) // Internal Server Error
		return
	}

	w.WriteHeader(http.StatusCreated) // 201 Created
	w.Write([]byte("Utilisateur creer"))
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pseudo   string `json:"pseudo"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var hash string
	var stats struct {
		TotalGames int
		Wins       int
		Losses     int
		Draws      int
	}

	err := data.DB.QueryRow(`SELECT password_hash, total_games, wins, losses, draws FROM users WHERE pseudo = ?`, req.Pseudo).Scan(&hash, &stats.TotalGames, &stats.Wins, &stats.Losses, &stats.Draws)
	if err != nil {
		http.Error(w, "Mot de passe ou pseudo incorrect", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"message":     "Connexion reussite",
		"total_games": stats.TotalGames,
		"wins":        stats.Wins,
		"losses":      stats.Losses,
		"draws":       stats.Draws,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpdateStats(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pseudo string `json:"pseudo"`
		Result string `json:"result"` // "win", "loss", or "draw"
	}
	json.NewDecoder(r.Body).Decode(&req)

	var query string
	switch req.Result {
	case "win":
		query = `UPDATE users SET wins = wins + 1, total_games = total_games + 1 WHERE pseudo = ?`
	case "loss":
		query = `UPDATE users SET losses = losses + 1, total_games = total_games + 1 WHERE pseudo = ?`
	case "draw":
		query = `UPDATE users SET draws = draws + 1, total_games = total_games + 1 WHERE pseudo = ?`
	default:
		http.Error(w, "Type de resultat invalide", http.StatusBadRequest)
		return
	}

	// update data Stats
	_, err := data.DB.Exec(query, req.Pseudo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur SQL : %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Statistiques updat"))
}

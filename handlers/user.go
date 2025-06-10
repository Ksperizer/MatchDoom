package handlers

import (
	"MatchDoom/data"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Pseudo   string `json:"pseudo"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	// Validation basique
	if req.Pseudo == "" || req.Password == "" || req.Email == "" {
		http.Error(w, "Tous les champs sont requis", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Erreur lors du hash", http.StatusInternalServerError)
		return
	}

	_, err = data.DB.Exec("INSERT INTO users (pseudo, password_hash, email) VALUES (?, ?, ?)", req.Pseudo, hash, req.Email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur SQL : %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Utilisateur créé avec succès",
	})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pseudo   string `json:"pseudo"`
		Password string `json:"password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	var hash string
	var stats struct {
		TotalGames int
		Wins       int
		Losses     int
		Draws      int
	}

	err := data.DB.QueryRow(`SELECT password_hash, total_games, wins, losses, draws FROM users WHERE pseudo = ?`, req.Pseudo).Scan(&hash, &stats.TotalGames, &stats.Wins, &stats.Losses, &stats.Draws)
	if err != nil {
		http.Error(w, "Pseudo ou mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	// Vérifier le mot de passe
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		http.Error(w, "Pseudo ou mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"message":     "Connexion réussie",
		"pseudo":      req.Pseudo,
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
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	var query string
	switch req.Result {
	case "win":
		query = `UPDATE users SET wins = wins + 1, total_games = total_games + 1 WHERE pseudo = ?`
	case "loss":
		query = `UPDATE users SET losses = losses + 1, total_games = total_games + 1 WHERE pseudo = ?`
	case "draw":
		query = `UPDATE users SET draws = draws + 1, total_games = total_games + 1 WHERE pseudo = ?`
	default:
		http.Error(w, "Type de résultat invalide (win/loss/draw)", http.StatusBadRequest)
		return
	}

	_, err := data.DB.Exec(query, req.Pseudo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur SQL : %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Statistiques mises à jour",
	})
}


func GenerateID() string {
	return fmt.Sprintf("web_client_%d", time.Now().UnixNano())
}
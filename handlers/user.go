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

type LoginRequest struct {
	Pseudo  string `json:"pseudo"`
	Password string `json:"password"`
}

type UpdateStatsRequest struct {
	Pseudo string `json:"pseudo"`
	Result string `json:"result"` // "win", "loss", or "draw"
}



func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	// validate required fields
	if req.Pseudo == "" || req.Password == "" || req.Email == "" {
		http.Error(w, "Tous les champs sont requis", http.StatusBadRequest)
		return
	}

	// check if pseudo and email are already used
	if existingUser, _ := data.GetUserByPseudo(req.Pseudo); existingUser != nil {
		http.Error(w, "Ce pseudo est déjà utilisé", http.StatusConflict)
		return
	}

	if existingUser, _ := data.GetUserByEmail(req.Email); existingUser != nil {
		http.Error(w, "Cet email est déjà utilisé", http.StatusConflict)
		return
	}

	// Hasher password 
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Erreur lors du hash", http.StatusInternalServerError)
		return
	}

	// Create user in the database
	err = data.CreateUser(req.Pseudo, string(hash), req.Email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur création utilisateur: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Utilisateur créé avec succès",
		"pseudo":  req.Pseudo,
	})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	// Récupérer l'utilisateur
	user, err := data.GetUserByPseudo(req.Pseudo)
	if err != nil {
		http.Error(w, "Pseudo ou mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	// check password hash 
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Pseudo ou mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	// return user information
	response := map[string]interface{}{
		"message":     "Connexion réussie",
		"pseudo":      user.Pseudo,
		"email":       user.Email,
		"user_id":     user.ID,
		"total_games": user.TotalGames,
		"wins":        user.Wins,
		"losses":      user.Losses,
		"draws":       user.Draws,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateStats update stats for a user
func UpdateStats(w http.ResponseWriter, r *http.Request) {
	var req UpdateStatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	
	user, err := data.GetUserByPseudo(req.Pseudo)
	if err != nil {
		http.Error(w, "Utilisateur non trouvé", http.StatusNotFound)
		return
	}


	err = data.UpdateUserStats(user.ID, req.Result)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur mise à jour stats: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Statistiques mises à jour",
	})
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
	pseudo := r.URL.Query().Get("pseudo")
	if pseudo == "" {
		http.Error(w, "Pseudo requis", http.StatusBadRequest)
		return
	}

	user, err := data.GetUserByPseudo(pseudo)
	if err != nil {
		http.Error(w, "Utilisateur non trouvé", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"id": 			user.ID,
		"pseudo": 		user.Pseudo,
		"email": 		user.Email,
		"created_at":	user.CreatedAt,
		"total_games":  user.TotalGames,
		"wins":         user.Wins,
		"losses":       user.Losses,
		"draws":        user.Draws,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLeaderboard retrieves the top 10 users based on their ranking
func GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	users, err := data.GetUserRanking(10)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur récupération classement: %v", err), http.StatusInternalServerError)
		return
	}

	var leaderboard []map[string]interface{}
	for i, user := range users {
		winRate := 0.0
		if user.TotalGames > 0 {
			winRate = float64(user.Wins) / float64(user.TotalGames) * 100
		}

		leaderboard = append(leaderboard, map[string]interface{}{
			"rank":        i + 1,
			"pseudo":      user.Pseudo,
			"total_games": user.TotalGames,
			"wins":        user.Wins,
			"losses":      user.Losses,
			"draws":       user.Draws,
			"win_rate":    fmt.Sprintf("%.1f%%", winRate),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"leaderboard": leaderboard,
		"total_players": len(leaderboard),
	})
}


func GetStats(w http.ResponseWriter, r *http.Request){
	stats, err := data.GetGameStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur récupération statistiques: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}


func JoinQueue(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pseudo string `json:"pseudo"`
		IP    string  `json:"ip"`
		Port  uint    `json:"port"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	_, err := data.GetUserByPseudo(req.Pseudo)
	if err != nil {
		http.Error(w, "Utilisateur non trouvé", http.StatusNotFound)
		return
	}

	err = data.AddToQueue(req.IP, req.Port, req.Pseudo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur ajout à la file d'attente: %v", err), http.StatusInternalServerError)
		return
	}

	count, _ := data.GetQueueCount()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Vous avez rejoint la file d'attente",
		"position": count,
	})
}


func GetActiveMatches(w http.ResponseWriter, r *http.Request) {
	matches, err := data.GetActiveMatches()
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur récupération parties: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"active_matches": matches,
		"count":         len(matches),
	})
}

func CreateTestData(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	err := data.PopulateTestData()
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur création données test: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Données de test créées avec succès",
	})
}

func GenerateID() string {
	return fmt.Sprintf("web_client_%d", time.Now().UnixNano())
}
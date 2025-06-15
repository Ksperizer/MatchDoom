package data 
import (
	"database/sql"
	"fmt"
	"time"


	_ "github.com/go-sql-driver/mysql" 
)

var DB *sql.DB

func InitDB() {
	dsn := "root:J&suisMySQL!1219@tcp(127.0.0.1:3306)/matchdoom?charset=utf8mb4&parseTime=true"
	
	var err error 
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	if err = DB.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("Connexion Mysql reussite")
	
	// Config pool connexion 
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(time.Hour)
}

type User struct {
	ID          uint      `json:"id"`
	Pseudo      string    `json:"pseudo"`
	PasswordHash string   `json:"password_hash"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	LastSeen    time.Time `json:"last_seen"`
	TotalGames  uint      `json:"total_games"`
	Wins        uint      `json:"wins"`
	Losses      uint      `json:"losses"`
	Draws       uint      `json:"draws"`
}


type QueueEntry struct {
	ID        uint      `json:"id"`
	IP        string    `json:"ip"`
	Port      uint      `json:"port"`
	Pseudo    string    `json:"pseudo"`
	CreatedAt time.Time `json:"created_at"`
}

type Match struct {
	ID         uint      `json:"id"`
	Player1ID  uint      `json:"player1_id"`
	Player2ID  uint      `json:"player2_id"`
	Board      string    `json:"board"`
	IsFinished bool      `json:"is_finished"`
	Winner     string    `json:"winner"`
	CreatedAt  time.Time `json:"created_at"`
}

type Move struct {
	ID        uint      `json:"id"`
	MatchID   uint      `json:"match_id"`
	Player    string    `json:"player"`
	Position  int       `json:"position"`
	PlayedAt  time.Time `json:"played_at"`
}


// Fonction TABLE USERS 
// CreateUser
func CreateUser(pseudo, passwordHash, email string) error {
	query := `INSERT INTO users (pseudo, password_hash, email) VALUES (?, ?, ?)`
	_, err := DB.Exec(query, pseudo, passwordHash, email)
	return err
}

// GetUserByPseudo 
func GetUserByPseudo(pseudo string) (*User, error){
	user := &User{}
	query := `SELECT id, pseudo, password_hash, email, created_at, total_games, wins, losses, draws FROM users WHERE pseudo = ?`
	
	err := DB.QueryRow(query, pseudo).Scan(
		&user.ID, &user.Pseudo, &user.PasswordHash, &user.Email, &user.CreatedAt,
		&user.TotalGames, &user.Wins, &user.Losses, &user.Draws,
	)

	if err != nil{
		return nil, err 
	}
	return user, nil
}

// GetUserByEmail
func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	query := `SELECT id, pseudo, password_hash, email, created_at, total_games, wins, losses, draws FROM users WHERE email = ?`
	
	err := DB.QueryRow(query, email).Scan(
		&user.ID, &user.Pseudo, &user.PasswordHash, &user.Email, &user.CreatedAt,
		&user.TotalGames, &user.Wins, &user.Losses, &user.Draws,
	)

	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(userID uint) (*User, error) {
	user := &User{}
	query := `SELECT id, pseudo, password_hash, email, created_at, total_games, wins, losses, draws FROM users WHERE id = ?`

	err := DB.QueryRow(query, userID).Scan(
		&user.ID, &user.Pseudo, &user.PasswordHash, &user.Email, &user.CreatedAt,
        &user.TotalGames, &user.Wins, &user.Losses, &user.Draws,
    )

	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserStats
func UpdateUserStats(userID uint, result string) error {
	var query string

	switch result {
	case "win":
		query = `UPDATE users SET wins = wins + 1, total_games = total_games + 1 WHERE id = ?`
	case "loss":
		query = `UPDATE users SET losses = losses + 1, total_games = total_games + 1 WHERE id = ?`
	case "draw":
		query = `UPDATE users SET draws = draws + 1, total_games = total_games + 1 WHERE id = ?`
	}

	_, err := DB.Exec(query, userID)
	return err
}

func GetAllUsers() ([]*User, error){
	query := `SELECT id, pseudo, password_hash, email, created_at, total_games, wins, losses, draws FROM users ORDER BY wins DESC`
	
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID, &user.Pseudo, &user.PasswordHash, &user.Email, &user.CreatedAt,
			&user.TotalGames, &user.Wins, &user.Losses, &user.Draws,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	
	return users, nil
}

// Fonction TABLE QUEUE

func AddToQueue(ip string, port uint, pseudo string) error {
	query := `INSERT INTO queue (ip, port, pseudo) VALUES (?, ?, ?)`
	_, err := DB.Exec(query, ip, port, pseudo)
	return err
}	


func GetQueueEntries() ([]*QueueEntry, error) {
	query := `SELECT id, ip, port, pseudo, created_at FROM queue ORDER BY created_at ASC`
	
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var entries []*QueueEntry
	for rows.Next() {
		entry := &QueueEntry{}
		err := rows.Scan(&entry.ID, &entry.IP, &entry.Port, &entry.Pseudo, &entry.CreatedAt)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	
	return entries, nil
}

func RemoveFromQueue(pseudo string) error {
	query := `DELETE FROM queue WHERE pseudo = ?`
	_, err := DB.Exec(query, pseudo)
	return err
}

func GetQueueCount() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM queue`
	err := DB.QueryRow(query).Scan(&count)
	return count, err
}

func CreateMatch(player1ID, player2ID uint) (uint, error) {
	query := `INSERT INTO matches (player1_id, player2_id) VALUES (?, ?)`
	
	result, err := DB.Exec(query, player1ID, player2ID)
	if err != nil {
		return 0, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	
	return uint(id), nil
}

// GetMatch  get match details by ID
func GetMatch(matchID uint) (*Match, error) {
	match := &Match{}
	query := `SELECT id, player1_id, player2_id, board, is_finished, winner, created_at FROM matches WHERE id = ?`
	
	err := DB.QueryRow(query, matchID).Scan(
		&match.ID, &match.Player1ID, &match.Player2ID, &match.Board,
		&match.IsFinished, &match.Winner, &match.CreatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	return match, nil
}

// UpdateMatchBoard update the board of a match
func UpdateMatchBoard(matchID uint, board string) error {
	query := `UPDATE matches SET board = ? WHERE id = ?`
	_, err := DB.Exec(query, board, matchID)
	return err
}

// FinishMatch finishes a match and updates the winner
func FinishMatch(matchID uint, winner string) error {
	query := `UPDATE matches SET is_finished = TRUE, winner = ? WHERE id = ?`
	_, err := DB.Exec(query, winner, matchID)
	
	if err != nil {
		return err
	}
	
	// update user stats based on the winner
	match, err := GetMatch(matchID)
	if err != nil {
		return err
	}
	
	switch winner {
	case "player1":
		UpdateUserStats(match.Player1ID, "win")
		UpdateUserStats(match.Player2ID, "loss")
	case "player2":
		UpdateUserStats(match.Player1ID, "loss")
		UpdateUserStats(match.Player2ID, "win")
	case "draw":
		UpdateUserStats(match.Player1ID, "draw")
		UpdateUserStats(match.Player2ID, "draw")
	}
	
	return nil
}


func GetActiveMatches() ([]*Match, error) {
	query := `SELECT id, player1_id, player2_id, board, is_finished, winner, created_at FROM matches WHERE is_finished = FALSE ORDER BY created_at DESC`
	
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var matches []*Match
	for rows.Next() {
		match := &Match{}
		err := rows.Scan(
			&match.ID, &match.Player1ID, &match.Player2ID, &match.Board,
			&match.IsFinished, &match.Winner, &match.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	
	return matches, nil
}

// GetAllMatches get all matches from the database
func GetAllMatches() ([]*Match, error) {
	query := `SELECT id, player1_id, player2_id, board, is_finished, winner, created_at FROM matches ORDER BY created_at DESC`
	
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var matches []*Match
	for rows.Next() {
		match := &Match{}
		err := rows.Scan(
			&match.ID, &match.Player1ID, &match.Player2ID, &match.Board,
			&match.IsFinished, &match.Winner, &match.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	
	return matches, nil
}


// AddMove 
func AddMove(matchID uint, player string, position int) error {
	query := `INSERT INTO moves (match_id, player, position) VALUES (?, ?, ?)`
	_, err := DB.Exec(query, matchID, player, position)
	return err
}

// GetMatchMoves 
func GetMatchMoves(matchID uint) ([]*Move, error) {
	query := `SELECT id, match_id, player, position, played_at FROM moves WHERE match_id = ? ORDER BY played_at ASC`
	
	rows, err := DB.Query(query, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var moves []*Move
	for rows.Next() {
		move := &Move{}
		err := rows.Scan(&move.ID, &move.MatchID, &move.Player, &move.Position, &move.PlayedAt)
		if err != nil {
			return nil, err
		}
		moves = append(moves, move)
	}
	
	return moves, nil
}

// GetLastMove retrieves the last move of a match
func GetLastMove(matchID uint) (*Move, error) {
	move := &Move{}
	query := `SELECT id, match_id, player, position, played_at FROM moves WHERE match_id = ? ORDER BY played_at DESC LIMIT 1`
	
	err := DB.QueryRow(query, matchID).Scan(
		&move.ID, &move.MatchID, &move.Player, &move.Position, &move.PlayedAt,
	)
	
	if err != nil {
		return nil, err
	}
	return move, nil
}

// FONCTIONS UTILITAIRES

// GetGameStats return stats for the game
func GetGameStats() (map[string]int, error) {
	stats := make(map[string]int)

	var totalUsers, activeMatches, totalMatches, queueCount int

	// Nombre total d'utilisateurs
	err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		return nil, err
	}

	
	err = DB.QueryRow("SELECT COUNT(*) FROM matches WHERE is_finished = FALSE").Scan(&activeMatches)
	if err != nil {
		return nil, err
	}

	
	err = DB.QueryRow("SELECT COUNT(*) FROM matches").Scan(&totalMatches)
	if err != nil {
		return nil, err
	}

	
	err = DB.QueryRow("SELECT COUNT(*) FROM queue").Scan(&queueCount)
	if err != nil {
		return nil, err
	}

	// Injection des résultats dans la map
	stats["total_users"] = totalUsers
	stats["active_matches"] = activeMatches
	stats["total_matches"] = totalMatches
	stats["queue_count"] = queueCount

	return stats, nil
}

// CleanOldQueue
func CleanOldQueue() error {
	query := `DELETE FROM queue WHERE created_at < DATE_SUB(NOW(), INTERVAL 1 HOUR)`
	_, err := DB.Exec(query)
	return err
}

// GetUserRanking 
func GetUserRanking(limit int) ([]*User, error) {
	query := `
		SELECT id, pseudo, password_hash, email, created_at, total_games, wins, losses, draws 
		FROM users 
		WHERE total_games > 0 
		ORDER BY wins DESC, total_games DESC 
		LIMIT ?
	`
	
	rows, err := DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID, &user.Pseudo, &user.PasswordHash, &user.Email, &user.CreatedAt,
			&user.TotalGames, &user.Wins, &user.Losses, &user.Draws,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	
	return users, nil
}

// UpdateUserLastSeen updates the last seen timestamp for a user
func UpdateUserLastSeen(userID uint) error {
	query := `UPDATE users SET last_seen = NOW() WHERE id = ?`
	_, err := DB.Exec(query, userID)
	return err 
}
	

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
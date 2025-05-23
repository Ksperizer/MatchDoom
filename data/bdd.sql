-- table queue -- 
CREATE TABLE queue (
    id INT AUTOINCREMENT PRIMARY KEY,
    ip varchar(55) NOT NULL,
    port INT NOT NULL,
    pseudo varchar(55) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
);

-- table matches --
CREATE TABLE matches (
    id INT PRIMARY KEY AUTOINCREMENT,
    player1_id INT NOT NULL,
    player2_id INT NOT NULL,
    board TEXT DEFAULT '',
    is_finised BOOLEAN  DEFAULT FALSE,
    winner ENUM('player1', 'player2', 'draw') DEFAULT NULL,
    CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (player1_id) REFERENCES users(id),
    FOREIGN KEY (player2_id) REFERENCES users(id)
);

-- table moves --
CREATE TABLE moves (
    id INT AUTOINCREMENT PRIMARY KEY,
    match_id INT NOT NULL,
    player ENUM('player1', 'player2') NOT NULL,
    position INT NOT NULL, -- 0-8 for 3x3 board
    player_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (match_id) REFERENCES matches(id) 
);


-- table users --
CREATE TABLE users (
    id INT PRIMARY KEY AUTOINCREMENT,
    pseudo VARCHAR(55) UNIQUE NOT NULL,
    password_hashed TEXT NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Stats --
    total_games INT DEFAULT 0,
    wins INT DEFAULT 0,
    losses INT DEFAULT 0,
    draws INT DEFAULT 0,
);
 
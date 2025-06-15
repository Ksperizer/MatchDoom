# 🎮 MatchDoom

> Jeu de morpion multijoueur temps réel avec interface Pygame, backend Go, logique réseau WebSocket et moteur Python.

## 🧠 Description

**MatchDoom** est un projet de jeu multijoueur en ligne où deux joueurs s'affrontent dans un **morpion** classique. L’infrastructure repose sur une architecture distribuée :  
- **Go** gère l’API REST, la file d’attente, les WebSockets côté client web.
- **Python** pilote la logique réseau pour le jeu en Pygame via WebSockets.
- **MySQL** stocke les utilisateurs, scores et classements.

---

## 🚀 Fonctionnalités principales

- 🔐 **Connexion / inscription** sécurisées (hash `bcrypt`, JSON API).
- 🕸️ **WebSocket matchmaking** temps réel (file d’attente, appairage).
- 🎲 **Interface Pygame** pour le jeu : intuitif, fluide, réactif.
- 📊 **Statistiques et classement** sauvegardés dans la base de données.
- 🌐 **Interface Web** (auth, stats, lancement de partie).

---

## 🛠️ Technologies utilisées

| Composant | Tech |
|----------|------|
| Backend | Go (Golang), Gorilla Mux, WebSocket |
| Frontend Web | HTML, JS, CSS |
| Jeu | Python 3, Pygame |
| Base de données | MySQL |
| Communication temps réel | WebSocket (Go ⇄ Python ⇄ Client) |

---

## 📁 Architecture

```
MatchDoom/
├── back/               # Backend Go (API, serveur WebSocket)
│   ├── server.go
│   ├── handlers/       # Handlers d'authentification et WebSocket
│   └── data/           # Connexion à MySQL
├── game/               # Moteur de jeu Python
│   ├── websocket.py    # Serveur WebSocket Python
│   └── ...             # Fichiers de logique de jeu
├── template/
│   ├── html/           # Pages HTML (accueil, connexion, profil, jeu)
│   ├── css/            # Styles CSS
│   └── script/         # JS pour appels API et WebSocket
└── README.md
```

## 🔧 Installation et Lancement

```
pip install websockets pygame
```

---

### 2. 🔙 Lancer le backend Go

```bash
go run main.go
```

- Serveur accessible sur: [http://localhost:8080/accueil](http://localhost:8080/accueil)
- WebSocket client : `ws://localhost:8080/game/ws`
- WebSocket Python (interne) : `ws://localhost:8081`

---

## 💡 Notes techniques

- Le matchmaking est géré côté Golang.
- Le bouton "Jouer" lance la connexion WebSocket côté client.
- Le backend Go redirige automatiquement vers les bonnes pages.


---

## 📚 Technologies

| Langage       | Usage                    |
|---------------|--------------------------|
| Go            | API REST, WebSocket      |
| Python        | Moteur de jeu temps réel |
| HTML/CSS/JS   | Frontend interactif      |
| MySQL         | Persistance données      |

---

## 🧑‍💻 Contributeurs

- **Cazeneuve Kévin** — Dev backend & intégration WebSocket
- **Harize Samson** — UI/UX, game logic, testing…

> Développé avec ❤️ pour un projet Ynov — 2025.  

---

## ✅ À faire / améliorations possibles

- 🎮 Affichage graphique du jeu dans la page HTML.
- 🧠 IA ou mode spectateur.
- 🌐 Matchmaking multijoueur élargi (files, rooms, etc.).

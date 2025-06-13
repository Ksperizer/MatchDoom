
# 🎮 MatchDoom - Jeu Tic Tac Toe Multijoueur avec Go, WebSocket & Python

Bienvenue sur **MatchDoom**, un projet de jeu en ligne de morpion (Tic Tac Toe) permettant à **deux joueurs** de s'affronter **en temps réel** grâce à **Go (backend/API)**, **HTML/CSS/JS (interface web)** et **Python (moteur de jeu via WebSocket)**.

---

## 🚀 Fonctionnalités

- 🔐 Connexion & inscription avec stockage sécurisé des utilisateurs (hash de mot de passe).
- 📊 Statistiques joueur (victoires, défaites, égalités, parties totales).
- 💬 Matchmaking temps réel via WebSocket (Go <=> Python).
- 🎨 Interface web moderne, responsive et interactive.
- 🧠 Moteur de jeu Python multijoueur (1v1) avec logique serveur.
- 💾 Base de données MySQL pour stocker les utilisateurs et matchs.

---

## 📁 Arborescence du projet

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

---

## 🛠️ Prérequis

- [Go](https://go.dev/) 
- [Python ](https://www.python.org/)
- [MySQL](https://www.mysql.com/) 
- Navigateur moderne (Chrome, Firefox…)

---

## 🔧 Installation et Lancement

### 1. ⚙️ Lancer le serveur Python (moteur de jeu)

```bash
cd game
python websocket.py
```

> Ce serveur écoute par défaut sur `ws://localhost:8081`

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

- Le matchmaking est géré côté Python.
- Le bouton "Jouer" lance la connexion WebSocket côté client.
- Le backend Go redirige automatiquement vers les bonnes pages.
- Les utilisateurs doivent être connectés (`localStorage.pseudo`) pour jouer.

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

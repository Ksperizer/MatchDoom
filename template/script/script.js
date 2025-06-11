document.addEventListener("DOMContentLoaded", () => {
// real time 
    function updateTime() {
        const now = new Date();
        const timeString = now.toLocateTimeString("fr-FR")
        const timeElement = document.getElementById("currentTime");
        if (timeElement) {
            timeElement.textContent = timeString;
        }   
    }

    updateTime(); // initial call
    setInterval(updateTime, 1000); // update every second

// Animate connexion status
    function updateConnectionStatus() {
        const statusDot = document.getELementById('apiStatus');
        const statusText = document.getElementById('statusText');
        const connectionStatus = document.getElementById('connectionStatus');

        if (!statusDot || !statusText || !connectionStatus) return;

        const status = [
            { class: '', text: 'initialisation...', status: 'En attente' },
            { class: 'connected', text: 'Connecté', status: 'Actif' },
            { class: 'Error', text: 'Erreur', status: 'Hors ligne' }
        ];

        let currentStatus = 0;

        setInterval(() => {
            currentStatus = (currentStatus + 1) % status.length;
            const status = statuses[currentStatus];
            
            statusDot.className = `status-dot ${status.class}`;
            statusText.textContent = status.text;
            connectionStatus.textContent = status.status;
        }, 3000); 
    }

    updateConnectionStatus();


    const refreshBtn = document.getElementById('refreshBtn');
    if (refreshBtn) {
        refreshBtn.addEventListener('click', () => {
            refreshBtn.style.transform = 'rotate(360deg)';
            setTimeout(() => {
                refreshBtn.style.transform = '';
                console.log('Actualisation...');
            }, 600);
        });
    }

    const settingsBtn = document.getElementById('settingsBtn');
    if (settingsBtn) {
        settingsBtn.addEventListener('click', () => {
            console.log('Paramètres');
        });
    }

    const playButton = document.getElementById('playButton');
    if (playButton) {
        playButton.addEventListener('click', () => {
            const pseudo = localStorage.getItem("pseudo");

            if (!pseudo) {
                alert("Vous devez vous connecter pour jouer !");
                window.location.href = "/connexion";
                return;
            }

            const ws = new WebSocket("ws://localhost:8080/game/ws");

            ws.onopen = () => {
                ws.send(JSON.stringify({ type: "join", pseudo }));
            };

            ws.onmessage = (event) => {
                const data = JSON.parse(event.data);

                if (data.type === "queue") {
                    alert(data.message + "\nPosition dans la file d'attente : " + data.position);
                }

                if (data.type === "game_start") {
                    alert(`Le jeu commence contre ${data.opponent} !`);
                    // Ici tu peux rediriger ou afficher une interface de jeu
                    localStorage.setItem("game_id", data.game_id);
                    localStorage.setItem("your_turn", data.your_turn);
                    localStorage.setItem("opponent", data.opponent);
                    localStorage.setItem("symbol", data.symbol);

                    new WebSocket("ws://localhost:8081"); // game python
                }
            };

            ws.onerror = (e) => {
                alert("Erreur WebSocket : " + e.message);
                console.error("WebSocket error:", e);
            };
        });
    }
});
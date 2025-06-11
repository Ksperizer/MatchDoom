document.addEventListener("DOMContentLoaded", () => {
    // ⏰ Show live time
    function updateTime() {
        const now = new Date();
        const timeString = now.toLocaleTimeString("fr-FR");
        const timeElement = document.getElementById("currentTime");
        if (timeElement) timeElement.textContent = timeString;
    }
    updateTime();
    setInterval(updateTime, 1000);

    // 🔄 Animate status connection (mock)
    function updateConnectionStatus() {
        const statusDot = document.getElementById('apiStatus');
        const statusText = document.getElementById('statusText');
        const connectionStatus = document.getElementById('connectionStatus');
        if (!statusDot || !statusText || !connectionStatus) return;

        const statuses = [
            { class: '', text: 'Initialisation...', status: 'En attente' },
            { class: 'connected', text: 'Connecté', status: 'Actif' },
            { class: 'error', text: 'Erreur', status: 'Hors ligne' }
        ];

        let currentStatus = 0;
        setInterval(() => {
            currentStatus = (currentStatus + 1) % statuses.length;
            const status = statuses[currentStatus];
            statusDot.className = `status-dot ${status.class}`;
            statusText.textContent = status.text;
            connectionStatus.textContent = status.status;
        }, 3000);
    }
    updateConnectionStatus();

    // 🔁 Refresh icon rotation
    const refreshBtn = document.getElementById('refreshBtn');
    if (refreshBtn) {
        refreshBtn.addEventListener('click', () => {
            refreshBtn.style.transform = 'rotate(360deg)';
            setTimeout(() => refreshBtn.style.transform = '', 600);
        });
    }

    // ⚙️ Placeholder settings button
    const settingsBtn = document.getElementById('settingsBtn');
    if (settingsBtn) {
        settingsBtn.addEventListener('click', () => console.log('Paramètres'));
    }

    // 🪟 Modal for login/register
    const authBtn = document.getElementById('authBtn');
    const profilBtn = document.getElementById('profilBtn');

    const modal = document.getElementById("authModal");

    modal.className = 'modal hidden';
    modal.innerHTML = `
        <div class="modal-content">
            <span class="close-btn" id="closeModal">&times;</span>
            <div class="tab">
                <button id="showLogin">Connexion</button>
                <button id="showRegister">Inscription</button>
            </div>
            <form id="loginForm" class="auth-form">
                <input type="text" name="pseudo" placeholder="Pseudo" required>
                <input type="password" name="password" placeholder="Mot de passe" required>
                <button type="submit">Se connecter</button>
            </form>
            <form id="registerForm" class="auth-form hidden">
                <input type="text" name="pseudo" placeholder="Pseudo" required>
                <input type="password" name="password" placeholder="Mot de passe" required>
                <input type="email" name="email" placeholder="Email" required>
                <button type="submit">S'inscrire</button>
            </form>
            <div id="authMessage" class="api-message"></div>
        </div>
    `;
    document.body.appendChild(modal);

    // Open modal
    authBtn.addEventListener("click", () => modal.classList.remove("hidden"));
    // Close modal
    document.getElementById("closeModal").onclick = () => modal.classList.add("hidden");

    // Toggle login/register
    document.getElementById("showLogin").onclick = () => {
        document.getElementById("loginForm").classList.remove("hidden");
        document.getElementById("registerForm").classList.add("hidden");
    };
    document.getElementById("showRegister").onclick = () => {
        document.getElementById("registerForm").classList.remove("hidden");
        document.getElementById("loginForm").classList.add("hidden");
    };

    // Login handler
    document.getElementById("loginForm").onsubmit = async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const res = await fetch("/api/login", {
            method: "POST",
            body: JSON.stringify(Object.fromEntries(formData)),
            headers: { "Content-Type": "application/json" }
        });
        const result = await res.json();
        document.getElementById("authMessage").textContent = result.message;
        if (res.ok) {
            localStorage.setItem("pseudo", formData.get("pseudo"));
            modal.classList.add("hidden");
            authBtn.classList.add("hidden");
            profilBtn.classList.remove("hidden");
        }
    };

    // Register handler
    document.getElementById("registerForm").onsubmit = async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const res = await fetch("/api/register", {
            method: "POST",
            body: JSON.stringify(Object.fromEntries(formData)),
            headers: { "Content-Type": "application/json" }
        });
        const result = await res.json();
        document.getElementById("authMessage").textContent = result.message;
    };

    // 🎮 Game start logic
    const playButton = document.getElementById('playButton');
    if (playButton) {
        playButton.addEventListener('click', () => {
            const pseudo = localStorage.getItem("pseudo");
            if (!pseudo) {
                alert("Vous devez vous connecter pour jouer !");
                return;
            }

            const ws = new WebSocket("ws://192.168.1.12:8081");

            ws.onopen = () => {
                ws.send(JSON.stringify({ type: "join", pseudo }));
            };

            ws.onmessage = (event) => {
                const data = JSON.parse(event.data);
                if (data.type === "queue") {
                    alert(`${data.message}\nPosition : ${data.position}`);
                } else if (data.type === "game_start") {
                    alert(`Le jeu commence contre ${data.opponent} !`);
                    localStorage.setItem("game_id", data.game_id);
                    localStorage.setItem("your_turn", data.your_turn);
                    localStorage.setItem("opponent", data.opponent);
                    localStorage.setItem("symbol", data.symbol);
                    window.location.href = "/game";
                }
            };

            ws.onerror = (e) => {
                alert("Erreur WebSocket : " + e.message);
                console.error("WebSocket error:", e);
            };
        });
    }
});

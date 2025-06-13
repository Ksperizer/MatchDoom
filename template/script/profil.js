// ===== CONFIGURATION =====
const CONFIG = {
    API_BASE_URL: "http://localhost:8080",
    WS_URL: "ws://localhost:8081",
    UPDATE_INTERVAL: 1000,
    STATUS_CYCLE_INTERVAL: 3000
};

// ===== VARIABLES GLOBALES =====
let isUserConnected = false;
let currentUser = null;
let statusInterval = null;
let timeInterval = null;

// ===== INITIALISATION =====
document.addEventListener("DOMContentLoaded", () => {
    console.log("🚀 Initialisation de MatchDoom...");
    
    initializeTime();
    initializeEventListeners();
    checkUserConnection();  // Déplacé après l'initialisation des listeners
    
    console.log("✅ MatchDoom initialisé avec succès!");
});

// ===== GESTION DE L'ÉTAT DE CONNEXION =====
function checkUserConnection() {
    const pseudo = localStorage.getItem("pseudo");
    
    if (pseudo) {
        console.log(`👤 Utilisateur connecté détecté: ${pseudo}`);
        setUserConnected(pseudo);
        loadUserStats();
    } else {
        console.log("🔒 Aucun utilisateur connecté");
        setUserDisconnected();
    }
}

function setUserConnected(pseudo) {
    isUserConnected = true;
    currentUser = pseudo;
    
    console.log(`🔄 Mise à jour interface pour: ${pseudo}`);
    
    // Mise à jour du header - utilise les IDs du HTML actuel
    const authBtn = document.getElementById("authBtn");
    const profilBtn = document.getElementById("profilBtn");
    const logoutBtn = document.getElementById("logoutBtn");
    
    if (authBtn) {
        authBtn.classList.add("hidden");
    }
    
    if (profilBtn) {
        profilBtn.classList.remove("hidden");
        profilBtn.textContent = `📊 Profil de ${pseudo}`;
        // Ajouter l'événement click pour le profil
        profilBtn.onclick = () => window.location.href = "/profil";
    }
    
    if (logoutBtn) {
        logoutBtn.classList.remove("hidden");
        logoutBtn.onclick = handleLogout;
    }
    
    // Mise à jour du contenu
    updateContentForConnectedUser(pseudo);
    
    // Fixer l'indicateur de statut sur "connecté"
    setStatusConnected();
    
    console.log(`✅ Interface mise à jour pour ${pseudo}`);
}

function setUserDisconnected() {
    isUserConnected = false;
    currentUser = null;
    
    console.log("🔄 Mise à jour interface pour visiteur");
    
    // Mise à jour du header
    const authBtn = document.getElementById("authBtn");
    const profilBtn = document.getElementById("profilBtn");
    const logoutBtn = document.getElementById("logoutBtn");
    
    if (authBtn) {
        authBtn.classList.remove("hidden");
        authBtn.textContent = "Connexion / Inscription";
    }
    
    if (profilBtn) {
        profilBtn.classList.add("hidden");
    }
    
    if (logoutBtn) {
        logoutBtn.classList.add("hidden");
    }
    
    // Mise à jour du contenu
    updateContentForGuest();
    
    // Redémarrer le cycle de statut normal
    initializeConnectionStatus();
}

function updateContentForConnectedUser(pseudo) {
    // Mettre à jour le titre et message de bienvenue
    const welcomeTitle = document.getElementById("welcomeTitle");
    const welcomeMessage = document.getElementById("welcomeMessage");
    const playButtonText = document.getElementById("playButtonText");
    
    if (welcomeTitle) {
        welcomeTitle.textContent = `Bienvenue, ${pseudo} !`;
    }
    
    if (welcomeMessage) {
        welcomeMessage.textContent = "Prêt pour une nouvelle partie de Tic-Tac-Toe ?";
    }
    
    if (playButtonText) {
        playButtonText.textContent = "Commencer une partie";
    }
}

function updateContentForGuest() {
    const welcomeTitle = document.getElementById("welcomeTitle");
    const welcomeMessage = document.getElementById("welcomeMessage");
    const playButtonText = document.getElementById("playButtonText");
    
    if (welcomeTitle) {
        welcomeTitle.textContent = "Bienvenue sur MatchDoom";
    }
    
    if (welcomeMessage) {
        welcomeMessage.textContent = "Connectez-vous pour défier d'autres joueurs au Tic-Tac-Toe !";
    }
    
    if (playButtonText) {
        playButtonText.textContent = "Se connecter pour jouer";
    }
}

function setStatusConnected() {
    // Arrêter le cycle automatique
    if (statusInterval) {
        clearInterval(statusInterval);
        statusInterval = null;
    }
    
    // Fixer le statut sur connecté
    const statusDot = document.getElementById('apiStatus');
    const statusText = document.getElementById('statusText');
    const connectionStatus = document.getElementById('connectionStatus');
    
    if (statusDot) statusDot.className = 'status-dot connected';
    if (statusText) statusText.textContent = 'Connecté';
    if (connectionStatus) connectionStatus.textContent = 'Actif';
    
    console.log("🟢 Statut fixé sur connecté");
}

// ===== CHARGEMENT DES STATISTIQUES =====
async function loadUserStats() {
    try {
        const response = await fetch(`${CONFIG.API_BASE_URL}/api/stats`);
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }
        
        const data = await response.json();
        console.log("📊 Statistiques chargées:", data);
        
        // Afficher les stats quelque part si nécessaire
        
    } catch (error) {
        console.error("❌ Erreur lors du chargement des stats:", error);
    }
}

// ===== GESTION DES ÉVÉNEMENTS =====
function initializeEventListeners() {
    // Bouton de connexion/inscription
    const authBtn = document.getElementById("authBtn");
    if (authBtn) {
        authBtn.addEventListener("click", openAuthModal);
    }
    
    // Bouton de jeu
    const playButton = document.getElementById("playButton");
    if (playButton) {
        playButton.addEventListener("click", startGame);
    }
    
    // Contrôles
    const refreshBtn = document.getElementById("refreshBtn");
    const settingsBtn = document.getElementById("settingsBtn");
    
    if (refreshBtn) refreshBtn.addEventListener("click", handleRefresh);
    if (settingsBtn) settingsBtn.addEventListener("click", () => console.log('Paramètres'));
    
    // Modal
    setupModalEventListeners();
    
    console.log("🎯 Event listeners initialisés");
}

function setupModalEventListeners() {
    const modal = document.getElementById("authModal");
    const closeModal = document.getElementById("closeModal");
    const showLogin = document.getElementById("showLogin");
    const showRegister = document.getElementById("showRegister");
    const loginForm = document.getElementById("loginForm");
    const registerForm = document.getElementById("registerForm");
    
    if (closeModal) closeModal.addEventListener("click", closeAuthModal);
    if (showLogin) showLogin.addEventListener("click", () => showAuthForm("login"));
    if (showRegister) showRegister.addEventListener("click", () => showAuthForm("register"));
    if (loginForm) loginForm.addEventListener("submit", handleLogin);
    if (registerForm) registerForm.addEventListener("submit", handleRegister);
    
    // Fermer modal en cliquant à l'extérieur (mais pas sur le contenu)
    if (modal) {
        modal.addEventListener("click", (e) => {
            if (e.target === modal) {
                closeAuthModal();
            }
        });
        
        // Empêcher la propagation des clics dans le contenu du modal
        const modalContent = modal.querySelector('.modal-content');
        if (modalContent) {
            modalContent.addEventListener("click", (e) => {
                e.stopPropagation();
            });
        }
    }
    
    // Forcer le focus sur les inputs quand ils sont cliqués
    const inputs = document.querySelectorAll('.auth-form input');
    inputs.forEach(input => {
        input.addEventListener('click', () => {
            input.focus();
        });
        
        input.addEventListener('touchstart', () => {
            input.focus();
        });
    });
}

// ===== GESTION DU MODAL D'AUTHENTIFICATION =====
function openAuthModal() {
    const modal = document.getElementById("authModal");
    if (modal) {
        modal.classList.remove("hidden");
        showAuthForm("login");
        
        // Forcer le focus sur le premier input après l'ouverture
        setTimeout(() => {
            const firstInput = modal.querySelector('.auth-form:not(.hidden) input[type="text"]');
            if (firstInput) {
                firstInput.focus();
                firstInput.click();
            }
        }, 100);
        
        console.log("🔐 Modal d'authentification ouvert");
    }
}

function closeAuthModal() {
    const modal = document.getElementById("authModal");
    const authMessage = document.getElementById("authMessage");
    
    if (modal) modal.classList.add("hidden");
    if (authMessage) authMessage.textContent = "";
    
    console.log("❌ Modal d'authentification fermé");
}

function showAuthForm(type) {
    const loginForm = document.getElementById("loginForm");
    const registerForm = document.getElementById("registerForm");
    const loginTab = document.getElementById("showLogin");
    const registerTab = document.getElementById("showRegister");
    const authMessage = document.getElementById("authMessage");
    
    if (type === "login") {
        if (loginForm) loginForm.classList.remove("hidden");
        if (registerForm) registerForm.classList.add("hidden");
        if (loginTab) loginTab.classList.add("active");
        if (registerTab) registerTab.classList.remove("active");
        
        // Focus sur le premier input du formulaire de connexion
        setTimeout(() => {
            const firstInput = loginForm?.querySelector('input[type="text"]');
            if (firstInput) {
                firstInput.focus();
            }
        }, 50);
    } else {
        if (registerForm) registerForm.classList.remove("hidden");
        if (loginForm) loginForm.classList.add("hidden");
        if (registerTab) registerTab.classList.add("active");
        if (loginTab) loginTab.classList.remove("active");
        
        // Focus sur le premier input du formulaire d'inscription
        setTimeout(() => {
            const firstInput = registerForm?.querySelector('input[type="text"]');
            if (firstInput) {
                firstInput.focus();
            }
        }, 50);
    }
    
    if (authMessage) authMessage.textContent = "";
}

// ===== GESTION DE L'AUTHENTIFICATION =====
async function handleLogin(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const messageBox = document.getElementById("authMessage");
    
    try {
        console.log("🔄 Tentative de connexion...");
        
        const response = await fetch(`${CONFIG.API_BASE_URL}/api/login`, {
            method: "POST",
            body: JSON.stringify(Object.fromEntries(formData)),
            headers: { "Content-Type": "application/json" }
        });
        
        if (response.ok) {
            const result = await response.json();
            localStorage.setItem("pseudo", result.pseudo);
            
            if (messageBox) {
                messageBox.textContent = "Connexion réussie !";
                messageBox.className = "api-message api-success";
            }
            
            console.log(`✅ Connexion réussie pour: ${result.pseudo}`);
            
            setTimeout(() => {
                closeAuthModal();
                // Actualiser la page pour appliquer les changements
                window.location.reload();
            }, 1000);
        } else {
            const errorText = await response.text();
            
            if (messageBox) {
                messageBox.textContent = errorText;
                messageBox.className = "api-message api-error";
            }
            
            console.error("❌ Erreur de connexion:", errorText);
        }
    } catch (error) {
        if (messageBox) {
            messageBox.textContent = "Erreur de connexion au serveur";
            messageBox.className = "api-message api-error";
        }
        
        console.error("❌ Erreur lors de la connexion:", error);
    }
}

async function handleRegister(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const messageBox = document.getElementById("authMessage");
    
    try {
        console.log("🔄 Tentative d'inscription...");
        
        const response = await fetch(`${CONFIG.API_BASE_URL}/api/register`, {
            method: "POST",
            body: JSON.stringify(Object.fromEntries(formData)),
            headers: { "Content-Type": "application/json" }
        });
        
        const result = await response.json();
        
        if (messageBox) {
            messageBox.textContent = result.message;
            messageBox.className = response.ok ? "api-message api-success" : "api-message api-error";
        }
        
        if (response.ok) {
            console.log("✅ Inscription réussie");
            setTimeout(() => showAuthForm("login"), 1500);
        } else {
            console.error("❌ Erreur d'inscription:", result.message);
        }
    } catch (error) {
        if (messageBox) {
            messageBox.textContent = "Erreur de connexion au serveur";
            messageBox.className = "api-message api-error";
        }
        
        console.error("❌ Erreur lors de l'inscription:", error);
    }
}

function handleLogout() {
    if (confirm("Êtes-vous sûr de vouloir vous déconnecter ?")) {
        localStorage.removeItem("pseudo");
        console.log("🚪 Utilisateur déconnecté");
        // Actualiser la page pour appliquer les changements
        window.location.reload();
    }
}

// ===== GESTION DU JEU =====
function startGame() {
    if (!isUserConnected) {
        console.log("🔒 Connexion requise pour jouer");
        openAuthModal();
        return;
    }
    
    console.log("🎮 Démarrage du jeu...");
    
    try {
        const ws = new WebSocket(CONFIG.WS_URL);
        
        ws.onopen = () => {
            console.log("🔗 Connexion WebSocket établie");
            ws.send(JSON.stringify({ type: "join", pseudo: currentUser }));
        };
        
        ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            console.log("📨 Message WebSocket reçu:", data);
            
            if (data.type === "queue") {
                alert(`${data.message}\nPosition : ${data.position}`);
            } else if (data.type === "game_start") {
                alert(`Le jeu commence contre ${data.opponent} !`);
                console.log("🎯 Partie démarrée!");
            } else if (data.type === "error") {
                alert(`Erreur: ${data.message}`);
            }
        };
        
        ws.onerror = (error) => {
            console.error("❌ Erreur WebSocket:", error);
            alert("Erreur: Impossible de se connecter au serveur de jeu");
        };
        
        ws.onclose = () => {
            console.log("🔌 Connexion WebSocket fermée");
        };
        
    } catch (error) {
        console.error("❌ Erreur lors du démarrage du jeu:", error);
        alert("Erreur: Impossible de démarrer le jeu");
    }
}

// ===== UTILITAIRES =====
function handleRefresh() {
    const refreshBtn = document.getElementById('refreshBtn');
    
    if (refreshBtn) {
        refreshBtn.style.transform = 'rotate(360deg)';
        setTimeout(() => refreshBtn.style.transform = '', 600);
    }
    
    console.log("🔄 Actualisation...");
    
    // Revérifier l'état de connexion
    checkUserConnection();
    
    if (isUserConnected) {
        loadUserStats();
    }
}

function initializeTime() {
    function updateTime() {
        const now = new Date();
        const timeString = now.toLocaleTimeString("fr-FR");
        const timeElement = document.getElementById("currentTime");
        if (timeElement) timeElement.textContent = timeString;
    }
    
    updateTime();
    timeInterval = setInterval(updateTime, CONFIG.UPDATE_INTERVAL);
    
    console.log("⏰ Horloge initialisée");
}

function initializeConnectionStatus() {
    // Seulement si l'utilisateur n'est pas connecté
    if (isUserConnected) {
        setStatusConnected();
        return;
    }
    
    function updateConnectionStatus() {
        const statusDot = document.getElementById('apiStatus');
        const statusText = document.getElementById('statusText');
        const connectionStatus = document.getElementById('connectionStatus');
        
        if (!statusDot || !statusText || !connectionStatus) return;

        const statuses = [
            { class: '', text: 'Initialisation...', status: 'En attente' },
            { class: 'connected', text: 'Prêt', status: 'Actif' },
            { class: 'error', text: 'Erreur', status: 'Hors ligne' }
        ];

        let currentStatus = 0;
        
        // Arrêter l'ancien interval s'il existe
        if (statusInterval) {
            clearInterval(statusInterval);
        }
        
        statusInterval = setInterval(() => {
            // Ne pas changer le statut si l'utilisateur est connecté
            if (isUserConnected) {
                setStatusConnected();
                return;
            }
            
            currentStatus = (currentStatus + 1) % statuses.length;
            const status = statuses[currentStatus];
            statusDot.className = `status-dot ${status.class}`;
            statusText.textContent = status.text;
            connectionStatus.textContent = status.status;
        }, CONFIG.STATUS_CYCLE_INTERVAL);
    }
    
    updateConnectionStatus();
    console.log("📡 Indicateur de statut initialisé");
}

// ===== NETTOYAGE =====
window.addEventListener('beforeunload', () => {
    if (timeInterval) clearInterval(timeInterval);
    if (statusInterval) clearInterval(statusInterval);
    console.log("🧹 Ressources nettoyées");
});

// ===== GESTION D'ERREURS GLOBALES =====
window.addEventListener('error', (event) => {
    console.error("❌ Erreur JavaScript:", event.error);
});

window.addEventListener('unhandledrejection', (event) => {
    console.error("❌ Promise rejetée:", event.reason);
});

console.log("📝 Script accueil.js chargé");
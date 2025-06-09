// ==========================================
// CONFIGURATION API - MODIFIEZ CES VALUES
// ==========================================
const API_BASE_URL = "http://localhost:8080"; // ⚠️ CHANGEZ L'URL DE VOTRE API ICI
const API_ENDPOINTS = {
  health: "/health", // ⚠️ AJOUTEZ VOS ENDPOINTS ICI
  // users: '/api/users',
  // data: '/api/data'
};

class APIManager {
  constructor() {
    this.baseUrl = API_BASE_URL;
    this.isConnected = false;
    this.init();
  }

  init() {
    this.setupEventListeners();
    this.startClock();
    this.checkConnection();

    // ⚠️ APPELEZ VOS FONCTIONS D'INITIALISATION ICI
    this.initializeAPI();
  }

  setupEventListeners() {
    // Boutons de contrôle
    document.getElementById("refreshBtn").addEventListener("click", () => {
      this.refreshAPI();
    });

    document.getElementById("settingsBtn").addEventListener("click", () => {
      this.showSettings();
    });
  }

  // ==========================================
  // SECTION À MODIFIER POUR VOTRE API
  // ==========================================

  /**
   * ⚠️ FONCTION PRINCIPALE - MODIFIEZ ICI POUR CHARGER VOTRE API
   * C'est ici que vous devez faire vos appels API
   */
  async initializeAPI() {
    try {
      this.showLoading();
      const health = await this.fetchAPI("/health");
      this.loadAPIContent();
    } catch (error) {
      this.showError(`Erreur d'initialisation: ${error.message}`);
    }
  }
     

  /**
   * ⚠️ FONCTION À MODIFIER - METTEZ À JOUR LE CONTENU DE VOTRE API ICI
   */
  async loadAPIContent() {
    try {
      // ⚠️ EXEMPLE - REMPLACEZ PAR VOTRE CONTENU RÉEL
      const apiContainer = document.getElementById("apiContainer");

      // Exemple de contenu dynamique depuis votre API Go
      const content = `
                <div class="api-success">
                    <h3>Connexion</h3>
                    <input type="text" id="pseudo" placeholder="Pseudo">
                    <input type="password" id="password" placeholder="Mot de passe">
                    <button onclick="window.apiManager.login()">Se connecter</button>
                </div>
            `;
      updateAPIContainer(content);
      this.updateStatus("connected", "API connectée");
    } catch (error) {
      this.showError(error.message);
    }
  }

  /**
   * ⚠️ FONCTION UTILITAIRE - UTILISEZ-LA POUR VOS APPELS API
   */
  async fetchAPI(endpoint, options = {}) {
    const url = `${this.baseUrl}${endpoint}`;
    const defaultOptions = {
      headers: this.getAPIHeaders(),
      ...options,
    };

    try {
      const response = await fetch(url, defaultOptions);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      // Essayer de parser en JSON, sinon retourner le texte
      const contentType = response.headers.get("content-type");
      if (contentType && contentType.includes("application/json")) {
        return await response.json();
      } else {
        return await response.text();
      }
    } catch (error) {
      console.error("Erreur API:", error);
      throw error;
    }
  }

  /**
   * ⚠️ MODIFIEZ LES HEADERS SI NÉCESSAIRE
   */
  getAPIHeaders() {
    return {
      "Content-Type": "application/json",
      Accept: "application/json",
      // ⚠️ AJOUTEZ VOS HEADERS D'AUTH ICI SI NÉCESSAIRE
      // 'Authorization': 'Bearer your-token'
    };
  }

  // ==========================================
  // FONCTIONS UTILITAIRES (ne pas modifier)
  // ==========================================

  async checkConnection() {
    try {
      // Essayer de se connecter à l'API
      await this.fetchAPI(API_ENDPOINTS.health);
      this.isConnected = true;
      this.updateStatus("connected", "Connecté");
    } catch (error) {
      this.isConnected = false;
      this.updateStatus("error", "Déconnecté");
      console.warn("API non disponible:", error.message);
    }
  }

  async login() {
    const pseudo = this.document.getElementById("pseudo").value;
    const password = this.document.getElementById("password").value;

    try {
      const response = await this.fetchAPI("/login", {
        method: "POST",
        body: JSON.stringify({ pseudo, password }),
      });

      this.updateStatus("connected", "Connect");
      updateAPIContainer(`<p>Bienvenue, ${pseudo}!</p>`);
    } catch (error) {
      this.showError("Echec de la connexion:" + error.message);
      }
    }

  updateStatus(status, text) {
    const statusDot = document.getElementById("apiStatus");
    const statusText = document.getElementById("statusText");
    const connectionStatus = document.getElementById("connectionStatus");

    statusDot.className = `status-dot ${status}`;
    statusText.textContent = text;
    connectionStatus.textContent = text;
  }

  showLoading() {
    const apiContainer = document.getElementById("apiContainer");
    apiContainer.innerHTML = `
            <div class="api-loading">
                <div class="loading-dots">
                    <span></span>
                    <span></span>
                    <span></span>
                </div>
                <p>Chargement de l'API...</p>
            </div>
        `;
  }

  showError(message) {
    const apiContainer = document.getElementById("apiContainer");
    apiContainer.innerHTML = `
            <div class="api-error">
                <h3>❌ Erreur</h3>
                <p>${message}</p>
                <button onclick="window.apiManager.refreshAPI()" 
                        style="margin-top: 15px; padding: 10px 20px; background: #dc2626; color: white; border: none; border-radius: 8px; cursor: pointer;">
                    Réessayer
                </button>
            </div>
        `;
    this.updateStatus("error", "Erreur");
  }

  refreshAPI() {
    this.initializeAPI();
  }

  showSettings() {
    alert(
      `Configuration API:\nURL: ${this.baseUrl}\nStatus: ${
        this.isConnected ? "Connecté" : "Déconnecté"
      }`
    );
  }

  startClock() {
    const updateTime = () => {
      const now = new Date();
      const timeString = now.toLocaleTimeString("fr-FR");
      document.getElementById("currentTime").textContent = timeString;
    };

    updateTime();
    setInterval(updateTime, 1000);
  }
}

// ==========================================
// INITIALISATION
// ==========================================
document.addEventListener("DOMContentLoaded", () => {
  // Créer l'instance globale du gestionnaire API
  window.apiManager = new APIManager();

  // ⚠️ AJOUTEZ VOS INITIALISATIONS SUPPLÉMENTAIRES ICI
});


function updateAPIContainer(content) {
  const container = document.getElementById("apiContainer");
  if (typeof content === "string") {
    container.innerHTML = content;
  } else {
    container.appendChild(content);
  }
}


function createElement(tag, className = "", content = "") {
  const element = document.createElement(tag);
  if (className) element.className = className;
  if (content) element.innerHTML = content;
  return element;
}


 // Fonction pour mettre à jour l'heure
        function updateTime() {
            const now = new Date();
            const timeString = now.toLocaleTimeString('fr-FR');
            const timeElement = document.getElementById('currentTime');
            if (timeElement) {
                timeElement.textContent = timeString;
            }
        }

        // Animation du statut de connexion
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

        // Interactions avec les boutons de contrôle
        document.addEventListener('DOMContentLoaded', function() {
            // Mettre à jour l'heure immédiatement et ensuite chaque seconde
            updateTime();
            setInterval(updateTime, 1000);
            
            // Démarrer l'animation du statut
            updateConnectionStatus();
            
            // Bouton refresh
            const refreshBtn = document.getElementById('refreshBtn');
            if (refreshBtn) {
                refreshBtn.addEventListener('click', function() {
                    this.style.transform = 'rotate(360deg)';
                    setTimeout(() => {
                        this.style.transform = '';
                        // Ici vous pouvez ajouter la logique de refresh de votre API
                        console.log('Actualisation...');
                    }, 600);
                });
            }
            
            // Bouton settings
            const settingsBtn = document.getElementById('settingsBtn');
            if (settingsBtn) {
                settingsBtn.addEventListener('click', function() {
                    console.log('Ouverture des paramètres...');
                    // Ici vous pouvez ajouter la logique des paramètres
                });
            }
        });
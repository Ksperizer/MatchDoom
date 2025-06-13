const API_BASE_URL = "http://localhost:8080";


function updateTime() {
    const now = new Date();
    const timeString = now.toLocaleTimeString("fr-FR");
    const timeElement = document.getElementById("currentTime");
    if (timeElement) timeElement.textContent = timeString;
}
updateTime();
setInterval(updateTime, 1000);

// Load profil 
async function loadProfile() {
    const pseudo = localStorage.getItem("pseudo");
    if (!pseudo) {
        window.location.href = "/accueil";
        return;
    }

    document.getElementById("sessionUser").textContent = pseudo;
    
    try {
        showLoading();
        
        const response = await fetch(`${API_BASE_URL}/api/profile?pseudo=${encodeURIComponent(pseudo)}`);
        
        if (!response.ok) {
            throw new Error(`Erreur ${response.status}: ${response.statusText}`);
        }
        
        const data = await response.json();
        displayProfile(data.user);
        
    } catch (error) {
        console.error("Erreur lors du chargement du profil:", error);
        showError(error.message);
    }
}

function showLoading() {
    document.getElementById("loadingIndicator").classList.remove("hidden");
    document.getElementById("profileData").classList.add("hidden");
    document.getElementById("errorMessage").classList.add("hidden");
}
A
function displayProfile(user) {
    document.getElementById("loadingIndicator").classList.add("hidden");
    document.getElementById("errorMessage").classList.add("hidden");
    document.getElementById("profileData").classList.remove("hidden");

    // update information 
    document.getElementById("pseudoDisplay").textContent = `Pseudo: ${user.pseudo}`;
    document.getElementById("emailDisplay").textContent = user.email;
    document.getElementById("totalGames").textContent = user.total_games;
    document.getElementById("wins").textContent = user.wins;
    document.getElementById("losses").textContent = user.losses;
    document.getElementById("draws").textContent = user.draws;

    // display created date
    const createdDate = new Date(user.created_at);
    document.getElementById("memberSince").textContent = createdDate.toLocaleDateString("fr-FR");

    // display win rate
    const winRate = user.total_games > 0 ? (user.wins / user.total_games * 100) : 0;
    document.getElementById("winRatePercent").textContent = `${winRate.toFixed(1)}%`;
    
    // update win rate circle
    const progressDegrees = (winRate / 100) * 360;
    const circle = document.querySelector(".win-rate-circle");
    circle.style.setProperty("--progress", `${progressDegrees}deg`);
}

function showError(message) {
    document.getElementById("loadingIndicator").classList.add("hidden");
    document.getElementById("profileData").classList.add("hidden");
    document.getElementById("errorMessage").classList.remove("hidden");
    document.getElementById("errorText").textContent = message;
}

// Classement
async function showLeaderboard() {
    document.getElementById("leaderboardModal").classList.remove("hidden");
    
    try {
        const response = await fetch(`${API_BASE_URL}/api/leaderboard?limit=10`);
        const data = await response.json();
        
        let tableHTML = `
            <table class="leaderboard-table">
                <thead>
                    <tr>
                        <th>Rang</th>
                        <th>Joueur</th>
                        <th>Victoires</th>
                        <th>Parties</th>
                        <th>Taux</th>
                    </tr>
                </thead>
                <tbody>
        `;
        
        data.leaderboard.forEach(player => {
            const medal = player.rank <= 3 ? ["", "🥇", "🥈", "🥉"][player.rank] : player.rank;
            tableHTML += `
                <tr>
                    <td><span class="rank-medal">${medal}</span></td>
                    <td>${player.pseudo}</td>
                    <td>${player.wins}</td>
                    <td>${player.total_games}</td>
                    <td>${player.win_rate}</td>
                </tr>
            `;
        });
        
        tableHTML += "</tbody></table>";
        document.getElementById("leaderboardContent").innerHTML = tableHTML;
        
    } catch (error) {
        document.getElementById("leaderboardContent").innerHTML = 
            '<div class="api-error">Erreur lors du chargement du classement</div>';
    }
}

function closeLeaderboard() {
    document.getElementById("leaderboardModal").classList.add("hidden");
}

// Event listeners
document.addEventListener("DOMContentLoaded", loadProfile);

document.getElementById("refreshProfileBtn").addEventListener("click", () => {
    document.getElementById("refreshProfileBtn").style.transform = 'rotate(360deg)';
    setTimeout(() => document.getElementById("refreshProfileBtn").style.transform = '', 600);
    loadProfile();
});

document.getElementById("logoutBtn").addEventListener("click", () => {
    localStorage.removeItem("pseudo");
    window.location.href = "/accueil";
});

document.getElementById("playFromProfileBtn").addEventListener("click", () => {
    window.location.href = "/accueil";
});

// Fermer le modal en cliquant à l'extérieur
document.getElementById("leaderboardModal").addEventListener("click", (e) => {
    if (e.target.classList.contains("modal")) {
        closeLeaderboard();
    }
});
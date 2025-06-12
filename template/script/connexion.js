document.addEventListener("DOMContentLoaded", () => {
    const loginForm = document.getElementById("loginForm");
    const registerForm = document.getElementById("registerForm");
    const messageBox = document.getElementById("authMessage");

    const showLoginBtn = document.getElementById("showLogin");
    const showRegisterBtn = document.getElementById("showRegister");

    if (!loginForm || !registerForm || !showLoginBtn || !showRegisterBtn) {
        console.warn("Un ou plusieurs éléments du formulaire de connexion ou d'inscription sont manquants.");
    }

    // display the login form by default
    showLoginBtn.onclick = () => {
        loginForm.classList.remove("hidden");
        registerForm.classList.add("hidden");
        messageBox.textContent = "";
    };

    // display the registration form
    showRegisterBtn.onclick = () => {
        registerForm.classList.remove("hidden");
        loginForm.classList.add("hidden");
        messageBox.textContent = "";
    };

    // submit login form
    loginForm.onsubmit = async (e) => {
        e.preventDefault();
        const formData = new FormData(loginForm);

        try {
            const res = await fetch("/api/login", {
                method: "POST",
                body: JSON.stringify(Object.fromEntries(formData)),
                headers: { "Content-Type": "application/json" }
            });

            if (res.ok) {
                const result = await res.json();
                localStorage.setItem("pseudo", result.pseudo); 
                messageBox.textContent = "Connexion réussie !";
                messageBox.className = "api-success";

                setTimeout(() => {
                    window.location.href = "/accueil"; 
                }, 1200);
            } else {
                const err = await res.text();
                messageBox.textContent = err;
                messageBox.className = "api-error";
            }
        } catch (error) {
            console.error("Erreur lors de la connexion :", error);
            messageBox.textContent = "Erreur serveur, réessayez plus tard.";
            messageBox.className = "api-error";
        }
    };

    // submit suscription form
    registerForm.onsubmit = async (e) => {
        e.preventDefault();
        const formData = new FormData(registerForm);

        try {
            const res = await fetch("/api/register", {
                method: "POST",
                body: JSON.stringify(Object.fromEntries(formData)),
                headers: { "Content-Type": "application/json" }
            });

            const result = await res.json();
            messageBox.textContent = result.message;
            messageBox.className = res.ok ? "api-success" : "api-error";

            if (res.ok) {
                // auto rempli 
                loginForm.pseudo.value = registerForm.pseudo.value;
                loginForm.password.value = "";
                showLoginBtn.click(); // bascule vers l’onglet connexion
            }

        } catch (error) {
            console.error("Erreur lors de l'inscription :", error);
            messageBox.textContent = "Erreur serveur, réessayez plus tard.";
            messageBox.className = "api-error";
        }
    };
});

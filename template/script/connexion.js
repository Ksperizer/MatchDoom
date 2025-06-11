const loginForm = document.getElementById('loginForm');
const registerForm = document.getElementById('registerForm');
const messageBox = document.getElementById("authMessage");

// display conditionnal forms
document.getElementById("showLogin").onclick = () => {
    loginForm.parentElement.classList.add("active");
    registerForm.parentElement.classList.remove("active");
    messageBox.innerText = "";
};

document.getElementById("showRegister").onclick = () => {
    registerForm.parentElement.classList.add("active");
    loginForm.parentElement.classList.remove("active");
    messageBox.innerText = "";
}

loginForm.onsubmit = async (e) => {
    e.preventDefault();
    const formData = new FormData(loginForm);

    const res = await fetch("/api/login", {
        method: "POST",
        body: JSON.stringify(Object.fromEntries(formData)),
        headers: { "Content-Type": "application/json" }
    });

    if (res.ok) {
        const result = await res.json();
        localStorage.setItem("pseudo", result.pseudo); // ⬅️ stock le pseudo
        messageBox.innerText = "Connexion réussie !";
        setTimeout(() => window.location.href = "/accueil", 1000);
    } else {
        const err = await res.text();
        messageBox.innerText = err;
    }
};

// send registration data
registerForm.onsubmit = async (e) => {
    e.preventDefault();
    const formData = new FormData(registerForm);
    const response = await fetch("/api/register", {
        method: "POST",
        body: JSON.stringify(Object.fromEntries(formData)),
        headers: { "Content-Type": "application/json" }
    });

    const result = await response.json();
    messageBox.innerText = result.message;
    messageBox.className = response.ok ? "api-success" : "api-error";
};
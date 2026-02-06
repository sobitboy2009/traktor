// Проверка сессии сразу при загрузке login.html
if (localStorage.getItem("isAuth") === "true") {
    window.location.href = "index.html"; // уже залогинен
}

document.getElementById("btnLogin").addEventListener("click", function() {
    const loginInput = document.getElementById("login").value;
    const passwordInput = document.getElementById("password").value;
    const errorDiv = document.getElementById("error");

    // Проверка логина и пароля
    if (loginInput === "Anvar" && passwordInput === "1727") {
        localStorage.setItem("isAuth", "true"); // сохраняем сессию
        window.location.href = "index.html";
    } else {
        errorDiv.textContent = "Login yoki parol noto‘g‘ri";
    }

});

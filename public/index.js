// ================================
// CONFIG
// ================================
const API_BASE = "/api";

// ================================
// DASHBOARD LOAD
// ================================
document.addEventListener("DOMContentLoaded", () => {
    loadDashboard();
});

// ================================
// LOAD DASHBOARD DATA
// ================================
async function loadDashboard() {
    try {
        const res = await fetch(`${API_BASE}/dashboard`);

        if (!res.ok) {
            throw new Error("Dashboard API error");
        }

        const data = await res.json();

        // SAFE UPDATE (если элементов нет — не падаем)
        const usersEl = document.getElementById("usersCount");
        const studentsEl = document.getElementById("studentsCount");
        const docsEl = document.getElementById("documentsCount");

        if (usersEl) usersEl.textContent = data.users ?? 0;
        if (studentsEl) studentsEl.textContent = data.students ?? 0;
        if (docsEl) docsEl.textContent = data.documents ?? 0;

        console.log("✅ Dashboard loaded:", data);

    } catch (err) {
        console.error("❌ Dashboard load failed:", err);

        // FALLBACK VALUES
        setZero("usersCount");
        setZero("studentsCount");
        setZero("documentsCount");
    }
}

// ================================
// HELPERS
// ================================
function setZero(id) {
    const el = document.getElementById(id);
    if (el) el.textContent = "0";
}

/* ================================
   STUDENTS (FOR OTHER PAGES)
================================ */

// GET ALL STUDENTS
async function getStudents() {
    const res = await fetch(`${API_BASE}/students`);
    return await res.json();
}

// GET ONE STUDENT
async function getStudent(jshshir) {
    const res = await fetch(`${API_BASE}/students/${jshshir}`);
    return await res.json();
}

// CREATE STUDENT
async function createStudent(data) {
    const res = await fetch(`${API_BASE}/students`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
    });

    return await res.json();
}

// UPDATE STUDENT
async function updateStudent(jshshir, data) {
    const res = await fetch(`${API_BASE}/students/${jshshir}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
    });

    return await res.json();
}

// DELETE STUDENT
async function deleteStudent(jshshir) {
    const res = await fetch(`${API_BASE}/students/${jshshir}`, {
        method: "DELETE",
    });

    return await res.json();
}

/* ================================
   DOCUMENTS
================================ */

async function getDocuments() {
    const res = await fetch(`${API_BASE}/documents`);
    return await res.json();
}

/* ================================
   INVOICES
================================ */

async function getInvoices() {
    const res = await fetch(`${API_BASE}/invoices`);
    return await res.json();
}

/* ================================
   DEBUG (НЕ УДАЛЯЙ)
================================ */
window.API = {
    getStudents,
    getStudent,
    createStudent,
    updateStudent,
    deleteStudent,
    getDocuments,
    getInvoices,
};


 

// Обработка клика по бургер-меню для показа/скрытия сайдбара
document.addEventListener('DOMContentLoaded', function() {
    const toggleSidebarBtn = document.querySelector('.toggle-sidebar-btn');
    const sidebar = document.getElementById('sidebar');
    const body = document.body;
    
    if (toggleSidebarBtn && sidebar) {
        toggleSidebarBtn.addEventListener('click', function(e) {
            e.preventDefault();
            body.classList.toggle('toggle-sidebar');
            
            // Также можно добавить/удалить класс непосредственно к сайдбару
            sidebar.classList.toggle('collapsed');
            
            // Сохраняем состояние в localStorage
            if (body.classList.contains('toggle-sidebar')) {
                localStorage.setItem('sidebarCollapsed', 'true');
            } else {
                localStorage.setItem('sidebarCollapsed', 'false');
            }
        });
        
        // Восстанавливаем состояние сайдбара при загрузке
        if (localStorage.getItem('sidebarCollapsed') === 'true') {
            body.classList.add('toggle-sidebar');
            sidebar.classList.add('collapsed');
        }
    }
    
    // Добавьте также этот код для автоматического сворачивания сайдбара на мобильных устройствах
    function handleResponsiveSidebar() {
        if (window.innerWidth <= 768) {
            body.classList.add('toggle-sidebar');
            sidebar.classList.add('collapsed');
        } else {
            // На десктопе восстанавливаем сохраненное состояние
            if (localStorage.getItem('sidebarCollapsed') !== 'true') {
                body.classList.remove('toggle-sidebar');
                sidebar.classList.remove('collapsed');
            }
        }
    }
    
    // Проверяем при загрузке и при изменении размера окна
    handleResponsiveSidebar();
    window.addEventListener('resize', handleResponsiveSidebar);
    
    







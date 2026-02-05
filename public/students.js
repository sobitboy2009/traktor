// students.js - Без alert, с фиксом кнопки
const API_BASE = "/api";


// ===== TOAST УВЕДОМЛЕНИЯ =====
function showToast(message, type = 'success') {
    // Создаем toast если его нет
    let toastId = type === 'success' ? 'successToast' : 'errorToast';
    let toastEl = document.getElementById(toastId);
    
    if (!toastEl) {
        // Создаем контейнер если его нет
        let container = document.querySelector('.toast-container');
        if (!container) {
            container = document.createElement('div');
            container.className = 'toast-container position-fixed bottom-0 end-0 p-3';
            document.body.appendChild(container);
        }
        
        // Создаем toast
        toastEl = document.createElement('div');
        toastEl.className = `toast align-items-center text-bg-${type === 'success' ? 'success' : 'danger'} border-0`;
        toastEl.id = toastId;
        toastEl.setAttribute('role', 'alert');
        toastEl.innerHTML = `
            <div class="d-flex">
                <div class="toast-body">
                    <i class="bi ${type === 'success' ? 'bi-check-circle' : 'bi-exclamation-triangle'} me-2"></i>
                    ${message}
                </div>
                <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
            </div>
        `;
        container.appendChild(toastEl);
    } else {
        // Обновляем сообщение
        const toastBody = toastEl.querySelector('.toast-body');
        if (toastBody) {
            toastBody.innerHTML = `
                <i class="bi ${type === 'success' ? 'bi-check-circle' : 'bi-exclamation-triangle'} me-2"></i>
                ${message}
            `;
        }
    }
    
    // Показываем toast
    const toast = new bootstrap.Toast(toastEl, {
        autohide: true,
        delay: 2000
    });
    toast.show();
    
    // Убираем toast после скрытия
    toastEl.addEventListener('hidden.bs.toast', function() {
        setTimeout(() => {
            if (toastEl && toastEl.parentNode) {
                toastEl.parentNode.removeChild(toastEl);
            }
        }, 100);
    });
}

// ===== УТИЛИТНЫЕ ФУНКЦИИ =====
function showLoading(show, elementId = null) {
    if (elementId) {
        const element = document.getElementById(elementId);
        if (element) {
            if (show) {
                element.setAttribute('data-original-text', element.innerHTML);
                element.innerHTML = '<span class="spinner-border spinner-border-sm me-2"></span>Yuklanmoqda...';
                element.disabled = true;
            } else {
                const originalText = element.getAttribute('data-original-text');
                if (originalText) {
                    element.innerHTML = originalText;
                }
                element.disabled = false;
            }
        }
    }
}

function formatDateForInput(dateString) {
    if (!dateString) return '';
    try {
        const date = new Date(dateString);
        if (isNaN(date.getTime())) return dateString;

        // Получаем год, месяц и день отдельно
        const year = date.getFullYear();
        // Месяцы в JS начинаются с 0, поэтому добавляем 1. 
        // padStart делает так, чтобы вместо "5" было "05"
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');

        // Собираем обратно в строку YYYY-MM-DD
        return `${year}-${month}-${day}`;
    } catch (e) {
        return dateString;
    }
}

function formatPhoneForDisplay(phone) {
    if (!phone) return '';
    
    // Убираем +998 если есть
    let cleanPhone = phone.replace(/^\+998/, '');
    cleanPhone = cleanPhone.replace(/\D/g, '');
    
    if (cleanPhone.length === 9) {
        return `${cleanPhone.substring(0, 2)} ${cleanPhone.substring(2, 5)}-${cleanPhone.substring(5, 7)}-${cleanPhone.substring(7)}`;
    }
    
    return phone;
}

function formatPhoneForAPI(phoneInput) {
    if (!phoneInput || phoneInput.trim() === '') return null;
    
    // Убираем все нецифровые символы
    let cleanPhone = phoneInput.replace(/\D/g, '');
    
    if (cleanPhone.length === 9) {
        return '+998' + cleanPhone;
    }
    
    if (cleanPhone.length === 12 && cleanPhone.startsWith('998')) {
        return '+' + cleanPhone;
    }
    
    return phoneInput;
}

function validateStudent(student) {
    // Проверка JShShIR
    if (!/^\d{14}$/.test(student.jshshir)) {
        showToast('JShShIR 14 ta raqamdan iborat bo\'lishi kerak', 'error');
        return false;
    }
    
    // Проверка ФИО
    if (!student.full_name || student.full_name.trim().length < 3) {
        showToast('To\'liq ism kamida 3 belgidan iborat bo\'lishi kerak', 'error');
        return false;
    }
    
    // Проверка даты рождения
    if (!student.birth_date) {
        showToast('Tug\'ilgan sana kiritilishi shart', 'error');
        return false;
    }
    
    const birthDate = new Date(student.birth_date);
    const today = new Date();
    if (birthDate > today) {
        showToast('Tug\'ilgan sana kelajakda bo\'lishi mumkin emas', 'error');
        return false;
    }
    
    return true;
}

// ===== ДЛЯ СТРАНИЦЫ РЕДАКТИРОВАНИЯ =====
function setupEditPage() {
    const params = new URLSearchParams(window.location.search);
    const jshshir = params.get('jshshir');
    
    if (!jshshir) {
        showToast('JShShIR parametri topilmadi!', 'error');
        setTimeout(() => window.location.href = 'students-list.html', 2000);
        return;
    }
    
    // Загружаем данные студента
    loadStudentForEdit(jshshir);
    
    // Настраиваем форму
    const form = document.getElementById('editStudentForm');
    if (form) {
        form.addEventListener('submit', function(e) {
            e.preventDefault();
            updateStudent(jshshir);
        });
        
        // Сохраняем оригинальный текст кнопки
        const submitBtn = form.querySelector('button[type="submit"]');
        if (submitBtn && !submitBtn.hasAttribute('data-original-text')) {
            submitBtn.setAttribute('data-original-text', submitBtn.innerHTML);
        }
    }
}

function loadStudentForEdit(jshshir) {
    // Показываем JShShIR сразу
    document.getElementById('editJshshirDisplay').textContent = jshshir;
    
    // Показываем состояние загрузки в кнопке
    const submitBtn = document.querySelector('#editStudentForm button[type="submit"]');
    if (submitBtn) {
        const originalText = submitBtn.innerHTML;
        submitBtn.setAttribute('data-original-text', originalText);
        submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm me-2"></span>Yuklanmoqda...';
        submitBtn.disabled = true;
    }
    
    fetch(`${API_URL}/students/${jshshir}`)
        .then(res => {
            if (!res.ok) {
                if (res.status === 404) {
                    throw new Error('Talaba topilmadi');
                }
                throw new Error('Server xatosi');
            }
            return res.json();
        })
        .then(student => {
            // Заполняем форму
            document.getElementById('editFullName').value = student.full_name || '';
            document.getElementById('editBirthDate').value = formatDateForInput(student.birth_date) || '';
            document.getElementById('editPhone').value = formatPhoneForDisplay(student.phone) || '';
            
            // Возвращаем кнопке нормальный вид
            if (submitBtn) {
                const originalText = submitBtn.getAttribute('data-original-text');
                submitBtn.innerHTML = originalText || '<i class="bi bi-check-circle me-2"></i>Saqlash';
                submitBtn.disabled = false;
            }
            
            // Тихий успех - без уведомления
            console.log('Talaba ma\'lumotlari yuklandi');
        })
        .catch(error => {
            console.error('Load error:', error);
            
            // Возвращаем кнопке нормальный вид
            if (submitBtn) {
                const originalText = submitBtn.getAttribute('data-original-text');
                submitBtn.innerHTML = originalText || '<i class="bi bi-check-circle me-2"></i>Saqlash';
                submitBtn.disabled = false;
            }
            
            showToast(`Talaba yuklanmadi: ${error.message}`, 'error');
            setTimeout(() => window.location.href = 'students-list.html', 3000);
        });
}

function updateStudent(jshshir) {
    const updatedStudent = {
        full_name: document.getElementById('editFullName').value.trim(),
        birth_date: document.getElementById('editBirthDate').value,
        phone: formatPhoneForAPI(document.getElementById('editPhone').value)
    };
    
    // Валидация
    if (!updatedStudent.full_name || updatedStudent.full_name.length < 3) {
        showToast('To\'liq ism kamida 3 belgidan iborat bo\'lishi kerak', 'error');
        return;
    }
    
    if (!updatedStudent.birth_date) {
        showToast('Tug\'ilgan sana kiritilishi shart', 'error');
        return;
    }
    
    // Меняем текст кнопки
    const submitBtn = document.querySelector('#editStudentForm button[type="submit"]');
    if (submitBtn) {
        const originalText = submitBtn.innerHTML;
        submitBtn.setAttribute('data-original-text', originalText);
        submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm me-2"></span>Saqlandi...';
        submitBtn.disabled = true;
    }
    
    fetch(`${API_URL}/students/${jshshir}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(updatedStudent)
    })
    .then(res => {
        if (!res.ok) {
            return res.text().then(text => {
                throw new Error(text || 'Server xatosi');
            });
        }
        return res.json();
    })
    .then(() => {
        // Успешное сохранение - показываем toast
        showToast('Talaba ma\'lumotlari muvaffaqiyatli yangilandi!', 'success');
        
        // Возвращаем кнопке нормальный вид
        if (submitBtn) {
            const originalText = submitBtn.getAttribute('data-original-text');
            submitBtn.innerHTML = originalText || '<i class="bi bi-check-circle me-2"></i>Saqlash';
            submitBtn.disabled = false;
        }
        
        // Перенаправление через 1.5 секунды
        setTimeout(() => {
            window.location.href = 'students-list.html';
        }, 1500);
    })
    .catch(error => {
        console.error('Update error:', error);
        showToast(`Yangilashda xatolik: ${error.message}`, 'error');
        
        // Возвращаем кнопке нормальный вид
        if (submitBtn) {
            const originalText = submitBtn.getAttribute('data-original-text');
            submitBtn.innerHTML = originalText || '<i class="bi bi-check-circle me-2"></i>Saqlash';
            submitBtn.disabled = false;
        }
    });
}

// ===== ДЛЯ СТРАНИЦЫ СПИСКА =====
function loadStudents() {
    const tbody = document.getElementById('students-table-body');
    if (!tbody) return;
    
    tbody.innerHTML = `
        <tr>
            <td colspan="6" class="text-center py-5">
                <div class="spinner-border text-primary"></div>
                <p class="mt-2">Yuklanmoqda...</p>
            </td>
        </tr>
    `;
    
    fetch(`${API_URL}/students`)
        .then(res => {
            if (!res.ok) throw new Error('Server xatosi');
            return res.json();
        })
        .then(students => {
            tbody.innerHTML = '';
            
            if (students.length === 0) {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="6" class="text-center py-5 text-muted">
                            <i class="bi bi-people display-4 d-block mb-3"></i>
                            Hech qanday o'quvchi topilmadi
                        </td>
                    </tr>
                `;
                return;
            }
            
            students.forEach((student, index) => {
                const row = document.createElement('tr');
                row.innerHTML = `
                    <td>${index + 1}</td>
                    <td><strong>${student.jshshir}</strong></td>
                    <td>${student.full_name}</td>
                    <td>${formatDateForInput(student.birth_date)}</td>
                    <td>${formatPhoneForDisplay(student.phone) || '-'}</td>
                    <td class="text-center">
                        <a href="students-view.html?jshshir=${student.jshshir}" 
                           class="btn btn-sm btn-primary me-1" title="Ko'rish">
                            <i class="bi bi-eye"></i>
                        </a>
                        <a href="students-edit.html?jshshir=${student.jshshir}" 
                           class="btn btn-sm btn-warning me-1" title="Tahrirlash">
                            <i class="bi bi-pencil"></i>
                        </a>
                        <button onclick="deleteStudent('${student.jshshir}')" 
                                class="btn btn-sm btn-danger" title="O'chirish">
                            <i class="bi bi-trash"></i>
                        </button>
                    </td>
                `;
                tbody.appendChild(row);
            });
            
            document.getElementById('totalStudents').textContent = students.length;
        })
        .catch(error => {
            console.error('Load students error:', error);
            tbody.innerHTML = `
                <tr>
                    <td colspan="6" class="text-center py-5 text-danger">
                        <i class="bi bi-exclamation-triangle display-4 d-block mb-3"></i>
                        Xatolik yuz berdi
                    </td>
                </tr>
            `;
        });
}

function deleteStudent(jshshir) {
    if (!confirm(`Rostan ham JShShIR: ${jshshir} bo'lgan o'quvchini o'chirmoqchimisiz?`)) {
        return;
    }
    
    fetch(`${API_URL}/students/${jshshir}`, {
        method: 'DELETE'
    })
    .then(res => {
        if (!res.ok) throw new Error('O\'chirishda xatolik');
        return res.json();
    })
    .then(() => {
        showToast('O\'quvchi muvaffaqiyatli o\'chirildi', 'success');
        loadStudents();
    })
    .catch(error => {
        console.error('Delete error:', error);
        showToast(`O'chirishda xatolik: ${error.message}`, 'error');
    });
}

// ===== ИНИЦИАЛИЗАЦИЯ =====
document.addEventListener('DOMContentLoaded', function() {
    // Определяем текущую страницу
    const path = window.location.pathname;
    
    if (path.includes('students-edit.html')) {
        setupEditPage();
    } else if (path.includes('students-list.html')) {
        loadStudents();
    }
});

// Экспорт функций
window.deleteStudent = deleteStudent;
window.loadStudents = loadStudents;





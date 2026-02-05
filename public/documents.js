// documents.js - Управление guvohnomalar (документами)
const API_BASE = "/api";

// Загрузка всех документов
async function loadDocuments() {
    try {
        const response = await fetch(`/api/documents`);
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Server error: ${response.status} - ${errorText}`);
        }
        
        const result = await response.json();
const documents = Array.isArray(result) ? result : result.data;

renderDocumentsTable(documents);
updateDocumentCount(documents);

    } catch (error) {
        console.error('Load documents error:', error);
        showError('Не удалось загрузить документы: ' + error.message);
    }
}


// Отображение таблицы документов
function renderDocumentsTable(documents) {
    const tbody = document.getElementById('certificates-table-body');
    if (!tbody) return;
    
    tbody.innerHTML = '';
    
    if (!documents || documents.length === 0) {
        tbody.innerHTML = `
            <tr>
                <td colspan="8" class="text-center py-5">
                    <i class="bi bi-file-earmark-x display-4 text-muted mb-3"></i>
                    <p class="h5">Guvohnomalar mavjud emas</p>
                    <a href="Guvohnomalar-add.html" class="btn btn-primary mt-2">
                        <i class="bi bi-plus-circle"></i> Birinchi guvohnomani yarating
                    </a>
                </td>
            </tr>
        `;
        return;
    }
    
    documents.forEach((doc, index) => {
        // Форматируем номер сертификата
        let certNumber = doc.certificate_number || '';
        
        // Удаляем префикс GUV- если есть
        if (certNumber.startsWith('GUV-')) {
            certNumber = certNumber.replace(/^GUV-/, '');
        }
        
        // Извлекаем только номер из формата ГГГГ-НОМЕР
        if (certNumber.includes('-')) {
            certNumber = certNumber.split('-').pop();
        }
        
        // Если нет номера, используем ID
        if (!certNumber && doc.id) {
            certNumber = doc.id.toString().padStart(6, '0');
        }
        
        const row = document.createElement('tr');
        row.innerHTML = `
            <td><strong>${index + 1}</strong></td>
            <td>
                <span class="badge-certificate">№${certNumber}</span>
            </td>
            <td>
                <div class="d-flex align-items-center">
                    <div class="avatar me-3">
                        <i class="bi bi-person-circle text-primary" style="font-size: 24px;"></i>
                    </div>
                    <div>
                        <h6 class="mb-0">${doc.student_name || 'Noma\'lum'}</h6>
                        <small class="text-muted">Talaba</small>
                    </div>
                </div>
            </td>
            <td><code>${doc.student_jshshir || 'Mavjud emas'}</code></td>
            <td>
                ${doc.categories ? doc.categories.split(',').map(cat => 
                    `<span class="badge bg-warning bg-gradient me-1">${cat.trim()}</span>`
                ).join('') : '<span class="badge bg-secondary">N/A</span>'}
            </td>
            <td>
                <span class="badge bg-info bg-gradient">
                    ${formatDate(doc.course_start)} - ${formatDate(doc.course_end)}
                </span>
            </td>
            <td>
                <span class="badge bg-success bg-gradient">
                    ${formatDate(doc.exam_date) || '-'}
                </span>
            </td>
            <td class="text-center">
                <div class="btn-group" role="group">
                    <a href="certificate.html?id=${doc.id}" class="btn btn-info btn-sm">
                        <i class="bi bi-eye"></i> Ko'rish
                    </a>
                    <button class="btn btn-sm btn-outline-warning" onclick="editCertificate(${doc.id})" title="Tahrirlash">
                        <i class="bi bi-pencil"></i>
                    </button>
                    <button class="btn btn-sm btn-outline-danger" onclick="deleteCertificate(${doc.id})" title="O'chirish">
                        <i class="bi bi-trash"></i>
                    </button>
                    <button class="btn btn-sm btn-outline-success" onclick="printCertificate(${doc.id})" title="Chop etish">
                        <i class="bi bi-printer"></i>
                    </button>
                </div>
            </td>
        `;
        tbody.appendChild(row);
    });
}

// Укажи здесь публичный URL своего сервера (ngrok, VPS или хостинг)
const PUBLIC_URL = 'https://traktor-production.up.railway.app/'; // <- замени на свой публичный адрес

function showQR(certNumber) {
    // Генерируем публичный URL для проверки сертификата
    const url = `${PUBLIC_URL}/verify.html?cert=${certNumber}`;

    const modalHtml = `
    <div class="modal fade" id="qrModal">
        <div class="modal-dialog modal-sm modal-dialog-centered">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">QR Code</h5>
                    <button class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body text-center">
                    <div id="qrcode"></div>
                    <p class="mt-2 small">${certNumber}</p>
                </div>
            </div>
        </div>
    </div>`;

    document.body.insertAdjacentHTML('beforeend', modalHtml);

    const modal = new bootstrap.Modal(document.getElementById('qrModal'));
    modal.show();

    new QRCode(document.getElementById("qrcode"), {
        text: url, // вот публичный URL
        width: 180,
        height: 180
    });

    document.getElementById('qrModal')
        .addEventListener('hidden.bs.modal', function () {
            this.remove();
        });
}



// Добавление нового документа
// Добавление нового документа
// Добавление нового документа
async function addDocument(event) {
    event.preventDefault();
    
    console.log('Starting document creation...');
    
    // Получаем данные из формы (правильные ID)
    const jshshir = document.getElementById('jshshir').value.trim();
    const studentName = document.getElementById('studentName').value.trim();
    
    // Получаем номер комиссии (из HTML поле с id="commissionNumber")
    const commissionNumberInput = document.getElementById('commissionNumber');
    const commissionNumber = commissionNumberInput ? commissionNumberInput.value.trim() : '';
    
    // Получаем имя директора (из двух полей в HTML)
    const directorNameInput = document.getElementById('directorName');
    const directorSurnameInput = document.getElementById('directorSurname');
    
    let directorName = '';
    if (directorNameInput && directorSurnameInput) {
        // Объединяем имя и фамилию директора
        directorName = `${directorNameInput.value.trim()} ${directorSurnameInput.value.trim()}`.trim();
    }
    
    // Проверки полей
    if (!jshshir || jshshir.length !== 14) {
        showError('JShShIR 14 raqamdan iborat bo\'lishi kerak!');
        return;
    }
    
    if (!studentName) {
        showError('Talaba ismini kiriting!');
        return;
    }
    
    if (!commissionNumber) {
        showError('Imtihon komissiyasi raqamini kiriting!');
        return;
    }
    
    // Получаем выбранные категории
    const categories = [];
    const categoryIds = ['categoryA', 'categoryB', 'categoryC', 'categoryD', 'categoryE', 'categoryF'];
    
    categoryIds.forEach(id => {
        const checkbox = document.getElementById(id);
        if (checkbox && checkbox.checked) {
            categories.push(checkbox.value);
        }
    });
    
    if (categories.length === 0) {
        showError('Kamida bitta toifani tanlang!');
        return;
    }
    
    // Получаем оценки
    const grade1 = document.querySelector('input[name="grade1"]:checked')?.value;
    const grade2 = document.querySelector('input[name="grade2"]:checked')?.value;
    
    if (!grade1 || !grade2) {
        showError('Barcha baholarni tanlang!');
        return;
    }
    
    // Собираем данные документа (ИСПРАВЛЕНО: правильные ID)
    const documentData = {
        title: "Traktor haydovchisi guvohnomasi",
        student_jshshir: jshshir,
        student_name: studentName,
        course_start: document.getElementById('courseStartDate').value,
        course_end: document.getElementById('courseEndDate').value,
        exam_date: document.getElementById('examDate').value,
        categories: categories.join(','),
        course_hours: parseInt(document.getElementById('courseHours').value) || 120,
        grade1: parseInt(grade1),
        grade2: parseInt(grade2),
        certificate_number: "",
        status: "active",
        // ИСПРАВЛЕНО: берем данные из формы
        commission_number: commissionNumber,  // Из поля с id="commissionNumber"
        director_name: directorName || 'N. ILYASOVA'  // Объединенное имя и фамилия
    };
    
    console.log('Sending document data:', documentData);
    
    try {
        const response = await fetch(`/api/documents`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(documentData)
        });
        
        const responseText = await response.text();
        console.log('Response status:', response.status);
        console.log('Response text:', responseText);
        
        if (!response.ok) {
            let errorMessage = 'Server error';
            try {
                const errorData = JSON.parse(responseText);
                errorMessage = errorData.message || errorData.error || responseText;
            } catch (e) {
                errorMessage = responseText || `HTTP ${response.status}`;
            }
            throw new Error(errorMessage);
        }
        
        const result = JSON.parse(responseText);
        showSuccess('✅ Guvohnoma muvaffaqiyatli qo\'shildi!');
        
        // Перенаправляем на список через 2 секунды
        setTimeout(() => {
            window.location.href = 'Guvohnomalar-list.html';
        }, 2000);
        
    } catch (error) {
        console.error('Add document error:', error);
        showError('❌ Guvohnoma qo\'shishda xatolik: ' + error.message);
    }
}

// Удаление документа
async function deleteCertificate(id) {
    if (!confirm(`Rostdan ham guvohnoma #${id} ni o'chirmoqchimisiz?\nBu amalni bekor qilib bo'lmaydi!`)) {
        return;
    }

    try {
        const response = await fetch(/api/documents/${id}`, {
            method: "DELETE"
        });

        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.message || `HTTP ${response.status}`);
        }

        showSuccess('✅ Guvohnoma muvaffaqiyatli o\'chirildi!');
        loadDocuments();

    } catch (error) {
        console.error("Delete certificate error:", error);
        showError(`❌ O'chirishda xatolik: ${error.message}`);
    }
}




// Редактирование документа (ОБЪЕДИНЕННАЯ ФУНКЦИЯ)
async function editCertificate(id) {
    window.location.href = `Guvohnomalar-edit.html?id=${id}`;
}

// Печать документа
function printCertificate(id) {
    // Временная заглушка для печати
    showSuccess(`Guvohnoma #${id} chop etish uchun tayyorlandi. Ctrl+P tugmalarini bosing.`);
    
    // Здесь можно добавить реальную логику печати
    setTimeout(() => {
        window.print();
    }, 1000);
}

// Вспомогательные функции
function formatDate(dateString) {
    if (!dateString) return '—';
    try {
        const date = new Date(dateString);
        const day = date.getDate().toString().padStart(2, '0');
        const month = (date.getMonth() + 1).toString().padStart(2, '0');
        const year = date.getFullYear();
        return `${day}.${month}.${year}`;
    } catch (e) {
        return dateString;
    }
}

function showError(message) {
    // Более красивое отображение ошибок
    const alertHtml = `
        <div class="alert alert-danger alert-dismissible fade show position-fixed top-0 start-50 translate-middle-x mt-3" style="z-index: 1060; min-width: 300px;" role="alert">
            <i class="bi bi-exclamation-triangle me-2"></i>
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        </div>
    `;
    
    document.body.insertAdjacentHTML('beforeend', alertHtml);
    
    // Автоматически закрыть через 5 секунд
    setTimeout(() => {
        const alert = document.querySelector('.alert-danger');
        if (alert) {
            const bsAlert = new bootstrap.Alert(alert);
            bsAlert.close();
        }
    }, 5000);
}

function showSuccess(message) {
    const alertHtml = `
        <div class="alert alert-success alert-dismissible fade show position-fixed top-0 start-50 translate-middle-x mt-3" style="z-index: 1060; min-width: 300px;" role="alert">
            <i class="bi bi-check-circle me-2"></i>
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        </div>
    `;
    
    document.body.insertAdjacentHTML('beforeend', alertHtml);
    
    // Автоматически закрыть через 3 секунды
    setTimeout(() => {
        const alert = document.querySelector('.alert-success');
        if (alert) {
            const bsAlert = new bootstrap.Alert(alert);
            bsAlert.close();
        }
    }, 3000);
}




function viewCertificate(id) {
    window.location.href = `certificate.html?id=${id}`;
}




function updateDocumentCount(documents) {
    document.getElementById('totalCertificates').textContent = documents.length;
    document.getElementById('totalCount').textContent = documents.length;

    const thisMonth = documents.filter(d => {
        if (!d.exam_date) return false;
        const date = new Date(d.exam_date);
        const now = new Date();
        return date.getMonth() === now.getMonth() &&
               date.getFullYear() === now.getFullYear();
    }).length;

    document.getElementById('thisMonth').textContent = thisMonth;

    document.getElementById('withCertificate').textContent =
        documents.filter(d => d.certificate_number).length;

    document.getElementById('printedToday').textContent = 0; // пока заглушка
}


// Инициализация
document.addEventListener('DOMContentLoaded', function() {
    // Загрузка документов для списка
    if (document.getElementById('certificates-table-body')) {
        loadDocuments();
    }
    
    // Обработка формы добавления документа
    const addForm = document.getElementById('addCertificateForm');
    if (addForm) {
        addForm.addEventListener('submit', addDocument);
    }
});

// Экспорт функций для использования в HTML
window.loadDocuments = loadDocuments;
window.deleteCertificate = deleteCertificate;
window.addDocument = addDocument;
window.viewCertificate = viewCertificate;
window.editCertificate = editCertificate;
window.printCertificate = printCertificate;
window.showQR = showQR;





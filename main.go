package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/skip2/go-qrcode"
	_ "github.com/lib/pq"
)

/* =========================
   GLOBALS
========================= */

var db *sql.DB

/* =========================
   MODELS
========================= */

type Dashboard struct {
	Users     int `json:"users"`
	Students  int `json:"students"`
	Documents int `json:"documents"`
}

type Student struct {
	JSHSHIR   string `json:"jshshir"`
	FullName  string `json:"full_name"`
	BirthDate string `json:"birth_date"`
	Phone     string `json:"phone"`
}

type Document struct {
	ID              int            `json:"id"`
	Title           sql.NullString `json:"title"`
	StudentJSHSHIR  sql.NullString `json:"student_jshshir"`
	StudentName     sql.NullString `json:"student_name"`
	CourseStart     sql.NullString `json:"course_start"`
	CourseEnd       sql.NullString `json:"course_end"`
	ExamDate        sql.NullString `json:"exam_date"`
	Categories      sql.NullString `json:"categories"`
	CourseHours     sql.NullInt64  `json:"course_hours"`
	Grade1          sql.NullInt64  `json:"grade1"`
	Grade2          sql.NullInt64  `json:"grade2"`
	CertificateNo   sql.NullString `json:"certificate_number"`
	Status          sql.NullString `json:"status"`
	CommissionNo    sql.NullString `json:"commission_number"`
	DirectorName    sql.NullString `json:"director_name"`
	CreatedAt       sql.NullString `json:"created_at"`
}

type DocumentOutput struct {
	ID              int    `json:"id"`
	Title           string `json:"title"`
	StudentJSHSHIR  string `json:"student_jshshir"`
	StudentName     string `json:"student_name"`
	CourseStart     string `json:"course_start"`
	CourseEnd       string `json:"course_end"`
	ExamDate        string `json:"exam_date"`
	Categories      string `json:"categories"`
	CourseHours     int    `json:"course_hours"`
	Grade1          int    `json:"grade1"`
	Grade2          int    `json:"grade2"`
	CertificateNo   string `json:"certificate_number"`
	Status          string `json:"status"`
	CommissionNo    string `json:"commission_number"`
	DirectorName    string `json:"director_name"`
	CreatedAt       string `json:"created_at"`
}

type DocumentDetail struct {
	DocumentOutput
	StudentBirthDate string `json:"student_birth_date"`
	StudentPhone     string `json:"student_phone"`
	QRCodeBase64     string `json:"qr_code_base64"`
}

type DocumentInput struct {
	Title           string `json:"title"`
	StudentJSHSHIR  string `json:"student_jshshir"`
	StudentName     string `json:"student_name"`
	CourseStart     string `json:"course_start"`
	CourseEnd       string `json:"course_end"`
	ExamDate        string `json:"exam_date"`
	Categories      string `json:"categories"`
	CourseHours     int    `json:"course_hours"`
	Grade1          int    `json:"grade1"`
	Grade2          int    `json:"grade2"`
	CertificateNo   string `json:"certificate_number"`
	Status          string `json:"status"`
	CommissionNo    string `json:"commission_number"`
	DirectorName    string `json:"director_name"`
}

type Invoice struct {
    ID              int       `json:"id"`
    StudentJSHSHIR  string    `json:"student_jshshir"`
    StudentName     string    `json:"student_name"`
    Description     string    `json:"description"`
    Amount          float64   `json:"amount"`
    Status          string    `json:"status"`
    InvoiceNumber   string    `json:"invoice_number"`
    CreatedAt       time.Time `json:"created_at"`
    IssueDate       string    `json:"issue_date,omitempty"`
    DueDate         string    `json:"due_date,omitempty"`
    PaymentDate     string    `json:"payment_date,omitempty"`
    StudentBirthDate string   `json:"student_birth_date,omitempty"`
    StudentPhone     string   `json:"student_phone,omitempty"`
}

type Certificate struct {
	ID               int    `json:"id"`
	StudentName      string `json:"student_name"`
	StudentJshshir   string `json:"student_jshshir"`
	Categories       string `json:"categories"`
	CourseStart      string `json:"course_start"`
	CourseEnd        string `json:"course_end"`
	ExamDate         string `json:"exam_date"`
	CourseHours      string `json:"course_hours"`
	Grade1           string `json:"grade1"`
	Grade2           string `json:"grade2"`
	CertificateNumber string `json:"certificate_number"`
	CommissionNumber string `json:"commission_number"`
}

/* =========================
   HELPERS
========================= */

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}



func convertDocumentToOutput(doc Document) DocumentOutput {
	return DocumentOutput{
		ID:              doc.ID,
		Title:           getStringValue(doc.Title),
		StudentJSHSHIR:  getStringValue(doc.StudentJSHSHIR),
		StudentName:     getStringValue(doc.StudentName),
		CourseStart:     getStringValue(doc.CourseStart),
		CourseEnd:       getStringValue(doc.CourseEnd),
		ExamDate:        getStringValue(doc.ExamDate),
		Categories:      getStringValue(doc.Categories),
		CourseHours:     int(getIntValue(doc.CourseHours)),
		Grade1:          int(getIntValue(doc.Grade1)),
		Grade2:          int(getIntValue(doc.Grade2)),
		CertificateNo:   getStringValue(doc.CertificateNo),
		Status:          getStringValue(doc.Status),
		CommissionNo:    getStringValue(doc.CommissionNo),
		DirectorName:    getStringValue(doc.DirectorName),
		CreatedAt:       getStringValue(doc.CreatedAt),
	}
}

func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func getIntValue(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}

func getNextCertificateNumber() (string, error) {
	// Ищем максимальный номер сертификата как число
	query := `
		SELECT MAX(CAST(certificate_number AS INTEGER)) 
		FROM documents 
		WHERE certificate_number ~ '^[0-9]+$'
	`
	
	var maxNumber sql.NullInt64
	err := db.QueryRow(query).Scan(&maxNumber)
	
	if err != nil {
		return "", err
	}
	
	// Если нет записей или ошибка, начинаем с 1
	if !maxNumber.Valid {
		return "0001", nil
	}
	
	nextNum := maxNumber.Int64 + 1
	return fmt.Sprintf("%04d", nextNum), nil
}

var BaseURL = "https://www.mttt-mexanizator.uz/" 

func generateQRCode(data string) (string, error) {
	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return "", err
	}
	
	var buf bytes.Buffer
	img := qr.Image(256)
	err = png.Encode(&buf, img)
	if err != nil {
		return "", err
	}
	
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

/* =========================
   DASHBOARD
========================= */

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	var d Dashboard

	db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&d.Users)
	db.QueryRow(`SELECT COUNT(*) FROM students`).Scan(&d.Students)
	db.QueryRow(`SELECT COUNT(*) FROM documents`).Scan(&d.Documents)

	respondJSON(w, d)
}

/* =========================
   STUDENTS
========================= */

func studentsList(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT jshshir, full_name, birth_date, phone FROM students`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var list []Student
	for rows.Next() {
		var s Student
		rows.Scan(&s.JSHSHIR, &s.FullName, &s.BirthDate, &s.Phone)
		list = append(list, s)
	}

	respondJSON(w, list)
}

func studentGet(w http.ResponseWriter, r *http.Request) {
	jshshir := mux.Vars(r)["jshshir"]

	var s Student
	err := db.QueryRow(`
		SELECT jshshir, full_name, birth_date, phone
		FROM students WHERE jshshir=$1`, jshshir,
	).Scan(&s.JSHSHIR, &s.FullName, &s.BirthDate, &s.Phone)

	if err != nil {
		http.Error(w, "Student not found", 404)
		return
	}

	respondJSON(w, s)
}

func studentCreate(w http.ResponseWriter, r *http.Request) {
	var s Student

	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		log.Println("❌ JSON decode error:", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`
		INSERT INTO students (jshshir, full_name, birth_date, phone)
		VALUES ($1,$2,$3,$4)
	`, s.JSHSHIR, s.FullName, s.BirthDate, s.Phone)

	if err != nil {
		log.Println("❌ INSERT student error:", err)
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, map[string]string{"status": "created"})
}


func studentUpdate(w http.ResponseWriter, r *http.Request) {
	jshshir := mux.Vars(r)["jshshir"]
	var s Student
	json.NewDecoder(r.Body).Decode(&s)

	_, err := db.Exec(`
		UPDATE students
		SET full_name=$1, birth_date=$2, phone=$3
		WHERE jshshir=$4`,
		s.FullName, s.BirthDate, s.Phone, jshshir,
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	respondJSON(w, map[string]string{"status": "updated"})
}

func studentDelete(w http.ResponseWriter, r *http.Request) {
	jshshir := mux.Vars(r)["jshshir"]

	db.Exec(`DELETE FROM students WHERE jshshir=$1`, jshshir)
	respondJSON(w, map[string]string{"status": "deleted"})
}

/* =========================
   DOCUMENTS
========================= */

func documentsList(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT id, title, student_jshshir, student_name,
		course_start, course_end, exam_date,
		categories, course_hours,
		grade1, grade2, certificate_number, status,
		commission_number, director_name, created_at
		FROM documents
		ORDER BY created_at DESC
	`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var docs []DocumentOutput
	for rows.Next() {
		var d Document
		err := rows.Scan(
			&d.ID, &d.Title, &d.StudentJSHSHIR, &d.StudentName,
			&d.CourseStart, &d.CourseEnd, &d.ExamDate,
			&d.Categories, &d.CourseHours,
			&d.Grade1, &d.Grade2,
			&d.CertificateNo, &d.Status,
			&d.CommissionNo, &d.DirectorName, &d.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning document: %v", err)
			continue
		}
		docs = append(docs, convertDocumentToOutput(d))
	}

	respondJSON(w, docs)
}

func documentGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", 400)
		return
	}
	
	var d Document
	err = db.QueryRow(`
		SELECT id, title, student_jshshir, student_name,
		course_start, course_end, exam_date,
		categories, course_hours,
		grade1, grade2, certificate_number, status,
		commission_number, director_name, created_at
		FROM documents WHERE id=$1`, id,
	).Scan(
		&d.ID, &d.Title, &d.StudentJSHSHIR, &d.StudentName,
		&d.CourseStart, &d.CourseEnd, &d.ExamDate,
		&d.Categories, &d.CourseHours,
		&d.Grade1, &d.Grade2,
		&d.CertificateNo, &d.Status,
		&d.CommissionNo, &d.DirectorName, &d.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Document not found", 404)
		} else {
			http.Error(w, err.Error(), 500)
		}
		return
	}

	respondJSON(w, convertDocumentToOutput(d))
}

func documentDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", 400)
		return
	}
	
	var detail DocumentDetail
	
	err = db.QueryRow(`
		SELECT 
			d.id, d.title, d.student_jshshir, d.student_name,
			d.course_start, d.course_end, d.exam_date,
			d.categories, d.course_hours,
			d.grade1, d.grade2, d.certificate_number, 
			d.status, d.commission_number, d.director_name, d.created_at,
			s.birth_date, s.phone
		FROM documents d
		LEFT JOIN students s ON d.student_jshshir = s.jshshir
		WHERE d.id = $1
	`, id).Scan(
		&detail.ID, &detail.Title, &detail.StudentJSHSHIR, &detail.StudentName,
		&detail.CourseStart, &detail.CourseEnd, &detail.ExamDate,
		&detail.Categories, &detail.CourseHours,
		&detail.Grade1, &detail.Grade2, &detail.CertificateNo,
		&detail.Status, &detail.CommissionNo, &detail.DirectorName, &detail.CreatedAt,
		&detail.StudentBirthDate, &detail.StudentPhone,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Document not found", 404)
		} else {
			http.Error(w, err.Error(), 500)
		}
		return
	}
	
	// Generate QR code
	qrData := fmt.Sprintf(
		"MMM TRAKTOR SERVIS\n" +
		"Guvohnoma: %s\n" +
		"Raqam: %d\n" +
		"Talaba: %s\n" +
		"JShShIR: %s\n" +
		"Sana: %s\n" +
		"Toifalar: %s\n" +
		"Imtihon: %s",
		detail.CertificateNo,
		detail.ID,
		detail.StudentName,
		detail.StudentJSHSHIR,
		detail.CreatedAt,
		detail.Categories,
		detail.ExamDate,
	)
	
	qrCode, err := generateQRCode(qrData)
	if err != nil {
		log.Printf("QR kod generatsiya xatosi: %v", err)
		detail.QRCodeBase64 = ""
	} else {
		detail.QRCodeBase64 = qrCode
	}
	
	respondJSON(w, detail)
}
isprav sam eto



















func documentCreate(w http.ResponseWriter, r *http.Request) {
	var input DocumentInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Printf("JSON dekodlash xatosi: %v", err)
		http.Error(w, "Noto'g'ri ma'lumot", 400)
		return
	}

	log.Printf("Qabul qilingan guvohnoma: %+v", input)

	// Установить значение по умолчанию для commission_number
	if input.CommissionNo == "" || strings.TrimSpace(input.CommissionNo) == "" {
		input.CommissionNo = "15"
		log.Printf("CommissionNo bo'sh, 15 ga o'rnatildi")
	}

	// Проверяем, существует ли студент
	if input.StudentJSHSHIR != "" && len(strings.TrimSpace(input.StudentJSHSHIR)) > 0 {
		var exists bool
		err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM students WHERE jshshir=$1)`, 
			strings.TrimSpace(input.StudentJSHSHIR)).Scan(&exists)
		if err != nil {
			log.Printf("Talaba mavjudligini tekshirish xatosi: %v", err)
			http.Error(w, "Baza xatosi", 500)
			return
		}
		
		if !exists {
			log.Printf("JShShIR %s bilan talaba topilmadi", input.StudentJSHSHIR)
			http.Error(w, "Talaba topilmadi", 404)
			return
		}
	}

	// Генерация номера сертификата (просто номер)
	if input.CertificateNo == "" || strings.TrimSpace(input.CertificateNo) == "" {
	certNumber, err := getNextCertificateNumber()
	if err != nil {
		log.Printf("Guvohnoma raqamini generatsiya qilish xatosi: %v", err)
		// В случае ошибки, используем ID как номер
		var maxID sql.NullInt64
		db.QueryRow("SELECT MAX(id) FROM documents").Scan(&maxID)
		nextNum := int64(1)
		if maxID.Valid {
			nextNum = maxID.Int64 + 1
		}
		certNumber = fmt.Sprintf("%04d", nextNum)
	}
	input.CertificateNo = certNumber
	log.Printf("Generatsiya qilingan guvohnoma raqami: %s", certNumber)
}

	// Вставка в базу данных
	_, err = db.Exec(`
		INSERT INTO documents 
		(title, student_jshshir, student_name, course_start, course_end, 
		 exam_date, categories, course_hours, grade1, grade2, 
		 certificate_number, status, commission_number, director_name, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())`,
		input.Title, input.StudentJSHSHIR, input.StudentName, input.CourseStart,
		input.CourseEnd, input.ExamDate, input.Categories, input.CourseHours,
		input.Grade1, input.Grade2, input.CertificateNo, input.Status,
		input.CommissionNo, input.DirectorName,
	)

	if err != nil {
		log.Printf("Guvohnoma yaratish xatosi: %v", err)
		http.Error(w, "Guvohnoma yaratishda xatolik: "+err.Error(), 500)
		return
	}

	respondJSON(w, map[string]interface{}{
		"status":            "success",
		"message":           "Guvohnoma muvaffaqiyatli yaratildi",
		"certificate_number": input.CertificateNo,
		"commission_number":  input.CommissionNo,
	})
}

// Добавьте этот handler в main.go
func verifyHandler(w http.ResponseWriter, r *http.Request) {
    cert := r.URL.Query().Get("cert")
    
    if cert == "" {
        http.Error(w, "Missing certificate parameter", 400)
        return
    }

    var doc DocumentOutput
    // Пробуем найти по номеру сертификата, если не найдено - по ID
    err := db.QueryRow(`
        SELECT id, certificate_number, student_name, student_jshshir,
               course_start, course_end, exam_date, categories,
               course_hours, grade1, grade2, status, director_name
        FROM documents 
        WHERE certificate_number=$1 OR id::text=$1
    `, cert).Scan(
        &doc.ID, &doc.CertificateNo, &doc.StudentName, &doc.StudentJSHSHIR,
        &doc.CourseStart, &doc.CourseEnd, &doc.ExamDate, &doc.Categories,
        &doc.CourseHours, &doc.Grade1, &doc.Grade2, &doc.Status, &doc.DirectorName,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Certificate not found", 404)
        } else {
            http.Error(w, err.Error(), 500)
        }
        return
    }

    respondJSON(w, doc)
}













func documentUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Noto'g'ri guvohnoma ID", 400)
		return
	}
	
	var input DocumentInput
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Noto'g'ri ma'lumot", 400)
		return
	}

	log.Printf("Yangilanayotgan guvohnoma ID %d ma'lumotlari: %+v", id, input)

	if input.StudentJSHSHIR != "" && len(strings.TrimSpace(input.StudentJSHSHIR)) > 0 {
		var exists bool
		err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM students WHERE jshshir=$1)`, 
			strings.TrimSpace(input.StudentJSHSHIR)).Scan(&exists)
		if err != nil {
			log.Printf("Talaba mavjudligini tekshirish xatosi: %v", err)
			http.Error(w, "Baza xatosi", 500)
			return
		}
		
		if !exists {
			log.Printf("JShShIR %s bilan talaba topilmadi", input.StudentJSHSHIR)
			http.Error(w, "Talaba topilmadi", 404)
			return
		}
	}

	result, err := db.Exec(`
		UPDATE documents 
		SET title=$1, student_jshshir=$2, student_name=$3, 
			course_start=$4, course_end=$5, exam_date=$6,
			categories=$7, course_hours=$8, grade1=$9, grade2=$10,
			certificate_number=$11, status=$12, 
			commission_number=$13, director_name=$14
		WHERE id=$15`,
		input.Title, input.StudentJSHSHIR, input.StudentName,
		input.CourseStart, input.CourseEnd, input.ExamDate,
		input.Categories, input.CourseHours, input.Grade1, input.Grade2,
		input.CertificateNo, input.Status,
		input.CommissionNo, input.DirectorName, id,
	)

	if err != nil {
		log.Printf("Guvohnoma yangilash xatosi: %v", err)
		http.Error(w, "Guvohnoma yangilashda xatolik: "+err.Error(), 500)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Guvohnoma topilmadi", 404)
		return
	}

	log.Printf("Guvohnoma ID %d muvaffaqiyatli yangilandi, ta'sirlangan qatorlar: %d", id, rowsAffected)

	respondJSON(w, map[string]interface{}{
		"status":        "success",
		"message":       "Guvohnoma muvaffaqiyatli yangilandi",
		"id":            id,
		"rows_affected": rowsAffected,
	})
}

func documentDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Noto'g'ri guvohnoma ID", 400)
		return
	}

	result, err := db.Exec(`DELETE FROM documents WHERE id=$1`, id)
	if err != nil {
		log.Printf("Guvohnoma o'chirish xatosi: %v", err)
		http.Error(w, "Guvohnoma o'chirishda xatolik: "+err.Error(), 500)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Guvohnoma topilmadi", 404)
		return
	}

	log.Printf("Guvohnoma ID %d muvaffaqiyatli o'chirildi", id)

	respondJSON(w, map[string]interface{}{
		"status":        "success",
		"message":       "Guvohnoma muvaffaqiyatli o'chirildi",
		"id":            id,
		"rows_affected": rowsAffected,
	})
}









/* =========================
   INVOICES
========================= */

func invoicesList(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query(`
        SELECT i.id, i.student_jshshir, 
               COALESCE(s.full_name, 'Noma''lum talaba') as student_name,
               i.description, i.amount, i.status, 
               COALESCE(i.invoice_number, 'INV-' || LPAD(i.id::text, 6, '0')) as invoice_number,
               i.created_at, i.issue_date, i.due_date, i.payment_date
        FROM invoices i
        LEFT JOIN students s ON i.student_jshshir = s.jshshir
        ORDER BY i.created_at DESC
    `)
    if err != nil {
        log.Printf("Error querying invoices: %v", err)
        http.Error(w, err.Error(), 500)
        return
    }
    defer rows.Close()

    var invoices []Invoice

    for rows.Next() {
        var i Invoice
        var issueDate, dueDate, paymentDate sql.NullString
        err := rows.Scan(
            &i.ID,
            &i.StudentJSHSHIR,
            &i.StudentName,
            &i.Description,
            &i.Amount,
            &i.Status,
            &i.InvoiceNumber,
            &i.CreatedAt,
            &issueDate,
            &dueDate,
            &paymentDate,
        )
        if err != nil {
            log.Printf("Error scanning invoice: %v", err)
            continue
        }
        invoices = append(invoices, i)
    }

    if err = rows.Err(); err != nil {
        log.Printf("Error iterating rows: %v", err)
        http.Error(w, err.Error(), 500)
        return
    }

    respondJSON(w, invoices)
}

func invoiceCreate(w http.ResponseWriter, r *http.Request) {
    var input struct {
        StudentJSHSHIR string  `json:"student_jshshir"`
        Description    string  `json:"description"`
        Amount         float64 `json:"amount"`
    }

    // Логируем полученные данные
    body, _ := io.ReadAll(r.Body)
    log.Printf("Received invoice create request: %s", string(body))
    
    // Восстанавливаем тело для парсинга
    r.Body = io.NopCloser(bytes.NewBuffer(body))
    
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        log.Printf("JSON decode error: %v", err)
        http.Error(w, "Invalid JSON: "+err.Error(), 400)
        return
    }

    log.Printf("Parsed data: JShShIR=%s, Description=%s, Amount=%f", 
        input.StudentJSHSHIR, input.Description, input.Amount)

    if input.StudentJSHSHIR == "" || input.Amount <= 0 {
        log.Printf("Missing fields: JShShIR='%s', Amount=%f", input.StudentJSHSHIR, input.Amount)
        http.Error(w, "Missing required fields", 400)
        return
    }

    // Получаем имя студента из базы
    var studentName string
    err := db.QueryRow(`SELECT full_name FROM students WHERE jshshir=$1`, 
        strings.TrimSpace(input.StudentJSHSHIR)).Scan(&studentName)
    
    if err != nil {
        if err == sql.ErrNoRows {
            log.Printf("Student not found with JShShIR: %s", input.StudentJSHSHIR)
            http.Error(w, "Talaba topilmadi. Avval talabani ro'yxatga oling.", 404)
            return
        }
        log.Printf("Error getting student: %v", err)
        studentName = "Noma'lum talaba"
    }

    log.Printf("Found student: %s", studentName)

    // Устанавливаем даты
    issueDate := time.Now().Format("2006-01-02")
    dueDate := time.Now().AddDate(0, 0, 30).Format("2006-01-02") // +30 дней
    
    var id int
    err = db.QueryRow(`
        INSERT INTO invoices (
            student_jshshir, student_name, description, amount, status,
            issue_date, due_date, created_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
        RETURNING id
    `,
        strings.TrimSpace(input.StudentJSHSHIR),
        studentName,
        input.Description,
        input.Amount,
        "To'lov kutilmoqda", // default status
        issueDate,
        dueDate,
    ).Scan(&id)

    if err != nil {
        log.Printf("Database error creating invoice: %v", err)
        http.Error(w, "Bazada xatolik: "+err.Error(), 500)
        return
    }

    // Генерируем номер инвойса
    invoiceNumber := fmt.Sprintf("INV-%06d", id)
    
    _, err = db.Exec(
        `UPDATE invoices SET invoice_number=$1 WHERE id=$2`,
        invoiceNumber, id,
    )

    if err != nil {
        log.Printf("Error updating invoice number: %v", err)
        // Не прерываем выполнение, так как инвойс уже создан
    }

    log.Printf("Invoice created successfully: ID=%d, Number=%s", id, invoiceNumber)

    respondJSON(w, map[string]interface{}{
        "success":        true,
        "id":            id,
        "invoice_number": invoiceNumber,
        "student_name":   studentName,
        "message":        "Invoyis muvaffaqiyatli yaratildi",
    })
}

func invoiceDelete(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    // Проверяем существование инвойса
    var exists bool
    err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM invoices WHERE id=$1)`, id).Scan(&exists)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    if !exists {
        http.Error(w, "Invoyis topilmadi", 404)
        return
    }

    res, err := db.Exec(`DELETE FROM invoices WHERE id=$1`, id)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    count, _ := res.RowsAffected()
    if count == 0 {
        http.Error(w, "Invoyis topilmadi", 404)
        return
    }

    respondJSON(w, map[string]string{
        "message": "Invoyis muvaffaqiyatli o'chirildi",
    })
}

func invoicesSearch(w http.ResponseWriter, r *http.Request) {
    q := "%" + r.URL.Query().Get("q") + "%"
    
    if q == "%%" {
        q = "%"
    }

    rows, err := db.Query(`
        SELECT i.id, i.student_jshshir, 
               COALESCE(s.full_name, 'Noma''lum talaba') as student_name,
               i.description, i.amount, i.status, 
               COALESCE(i.invoice_number, 'INV-' || LPAD(i.id::text, 6, '0')) as invoice_number,
               i.created_at
        FROM invoices i
        LEFT JOIN students s ON i.student_jshshir = s.jshshir
        WHERE i.student_jshshir ILIKE $1
           OR s.full_name ILIKE $1
           OR i.description ILIKE $1
           OR i.invoice_number ILIKE $1
        ORDER BY i.created_at DESC
    `, q)

    if err != nil {
        log.Printf("Search error: %v", err)
        http.Error(w, err.Error(), 500)
        return
    }
    defer rows.Close()

    var invoices []Invoice
    for rows.Next() {
        var i Invoice
        err := rows.Scan(
            &i.ID,
            &i.StudentJSHSHIR,
            &i.StudentName,
            &i.Description,
            &i.Amount,
            &i.Status,
            &i.InvoiceNumber,
            &i.CreatedAt,
        )
        if err != nil {
            log.Printf("Error scanning search result: %v", err)
            continue
        }
        invoices = append(invoices, i)
    }

    respondJSON(w, invoices)
}

// Функция для обновления статуса инвойса
func invoiceUpdateStatus(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    
    // Проверяем ID
    invoiceID, err := strconv.Atoi(id)
    if err != nil {
        http.Error(w, "Invalid invoice ID", 400)
        return
    }
    
    var input struct {
        Status string `json:"status"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid JSON", 400)
        return
    }
    
    // Проверяем допустимые статусы
    validStatuses := []string{"To'lov kutilmoqda", "To'landi", "Bekor qilindi"}
    isValid := false
    for _, status := range validStatuses {
        if input.Status == status {
            isValid = true
            break
        }
    }
    
    if !isValid {
        http.Error(w, "Invalid status", 400)
        return
    }
    
    // Обновляем статус и дату оплаты если статус "To'landi"
    var paymentDate interface{}
    if input.Status == "To'landi" {
        paymentDate = time.Now().Format("2006-01-02")
    } else {
        paymentDate = nil
    }
    
    result, err := db.Exec(`
        UPDATE invoices 
        SET status = $1, payment_date = $2 
        WHERE id = $3
    `, input.Status, paymentDate, invoiceID)
    
    if err != nil {
        log.Printf("Error updating invoice status: %v", err)
        http.Error(w, err.Error(), 500)
        return
    }
    
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        http.Error(w, "Invoice not found", 404)
        return
    }
    
    respondJSON(w, map[string]interface{}{
        "success": true,
        "message": "Invoyis holati yangilandi",
        "status": input.Status,
    })
}

// Функция для получения деталей инвойса (если нужна)

func invoiceGetDetails(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    
    invoiceID, err := strconv.Atoi(id)
    if err != nil {
        http.Error(w, "Invalid invoice ID", 400)
        return
    }
    
    var invoiceDetail struct {
        ID              int            `json:"id"`
        StudentJSHSHIR  string         `json:"student_jshshir"`
        StudentName     string         `json:"student_name"`
        Description     string         `json:"description"`
        Amount          float64        `json:"amount"`
        Status          string         `json:"status"`
        InvoiceNumber   string         `json:"invoice_number"`
        IssueDate       string         `json:"issue_date"`
        DueDate         string         `json:"due_date"`
        PaymentDate     string         `json:"payment_date,omitempty"`
        CreatedAt       string         `json:"created_at"`
        StudentBirthDate string        `json:"student_birth_date,omitempty"`
        StudentPhone     string        `json:"student_phone,omitempty"`
    }
    
    var issueDate, dueDate, paymentDate, studentBirthDate, studentPhone sql.NullString
    
    err = db.QueryRow(`
        SELECT 
            i.id, i.student_jshshir, i.student_name,
            i.description, i.amount, i.status, 
            COALESCE(i.invoice_number, 'INV-' || LPAD(i.id::text, 6, '0')) as invoice_number,
            i.issue_date, i.due_date, i.payment_date,
            i.created_at,
            s.birth_date, s.phone
        FROM invoices i
        LEFT JOIN students s ON i.student_jshshir = s.jshshir
        WHERE i.id = $1
    `, invoiceID).Scan(
        &invoiceDetail.ID,
        &invoiceDetail.StudentJSHSHIR,
        &invoiceDetail.StudentName,
        &invoiceDetail.Description,
        &invoiceDetail.Amount,
        &invoiceDetail.Status,
        &invoiceDetail.InvoiceNumber,
        &issueDate,
        &dueDate,
        &paymentDate,
        &invoiceDetail.CreatedAt,
        &studentBirthDate,
        &studentPhone,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Invoice not found", 404)
        } else {
            log.Printf("Error getting invoice details: %v", err)
            http.Error(w, err.Error(), 500)
        }
        return
    }
    
    // Преобразуем NullString в обычные строки
    if issueDate.Valid {
        invoiceDetail.IssueDate = issueDate.String
    }
    if dueDate.Valid {
        invoiceDetail.DueDate = dueDate.String
    }
    if paymentDate.Valid {
        invoiceDetail.PaymentDate = paymentDate.String
    }
    if studentBirthDate.Valid {
        invoiceDetail.StudentBirthDate = studentBirthDate.String
    }
    if studentPhone.Valid {
        invoiceDetail.StudentPhone = studentPhone.String
    }
    
    respondJSON(w, invoiceDetail)
}






/* =========================
   MAIN - ИСПРАВЛЕННАЯ ВЕРСИЯ
========================= */

func main() {
  var err error

  // Подключение к базе данных
  dbURL := os.Getenv("DATABASE_URL")
if dbURL == "" {
  log.Fatal("❌ DATABASE_URL не задана")
}

db, err = sql.Open("postgres", dbURL)
if err != nil {
  log.Fatal(err)
}

  // Настройка пула соединений
  db.SetMaxOpenConns(25)
  db.SetMaxIdleConns(10)
  db.SetConnMaxLifetime(5 * time.Minute)

  // Проверка соединения
  err = db.Ping()
  if err != nil {
    log.Fatal("BAZA XATOSI:", err)
  }

  log.Println("✅ Baza ulandi")

  // Создание роутера
  r := mux.NewRouter()

  // API маршруты
  r.HandleFunc("/api/dashboard", enableCORS(dashboardHandler)).Methods("GET")
  
  // Students API
  r.HandleFunc("/api/students", enableCORS(studentsList)).Methods("GET")
  r.HandleFunc("/api/students", enableCORS(studentCreate)).Methods("POST")
  r.HandleFunc("/api/students/{jshshir}", enableCORS(studentGet)).Methods("GET")
  r.HandleFunc("/api/students/{jshshir}", enableCORS(studentUpdate)).Methods("PUT")
  r.HandleFunc("/api/students/{jshshir}", enableCORS(studentDelete)).Methods("DELETE")
  
  // Documents API
  r.HandleFunc("/api/documents", enableCORS(documentsList)).Methods("GET")
  r.HandleFunc("/api/documents", enableCORS(documentCreate)).Methods("POST")
  r.HandleFunc("/api/documents/{id}", enableCORS(documentGet)).Methods("GET")
  r.HandleFunc("/api/documents/{id}/details", enableCORS(documentDetails)).Methods("GET")
  r.HandleFunc("/api/documents/{id}", enableCORS(documentUpdate)).Methods("PUT")
  r.HandleFunc("/api/documents/{id}", enableCORS(documentDelete)).Methods("DELETE")


  r.HandleFunc("/api/invoices", enableCORS(invoicesList)).Methods("GET")
r.HandleFunc("/api/invoices", enableCORS(invoiceCreate)).Methods("POST")
r.HandleFunc("/api/invoices/{id}", enableCORS(invoiceDelete)).Methods("DELETE")
r.HandleFunc("/api/invoices/search", enableCORS(invoicesSearch)).Methods("GET")
r.HandleFunc("/api/invoices/{id}/details", enableCORS(invoiceGetDetails)).Methods("GET")
r.HandleFunc("/api/invoices/{id}/status", enableCORS(invoiceUpdateStatus)).Methods("PUT")

  // ВАЖНОЕ ИСПРАВЛЕНИЕ: Путь к статическим файлам
  // Получаем текущую директорию
  currentDir, err := os.Getwd()
  if err != nil {
    log.Fatal("Direktoriyani o'qib bo'lmadi:", err)
  }
  
  // Проверяем, находимся ли мы в папке backend
  publicPath := ""
  if strings.HasSuffix(currentDir, "backend") {
    // Если в backend, то public на уровень выше
    publicPath = filepath.Join(filepath.Dir(currentDir), "public")
  } else {
    // Иначе предполагаем, что мы в корне проекта
    publicPath = filepath.Join(currentDir, "public")
  }
  
  // Проверяем существование папки public
  if _, err := os.Stat(publicPath); os.IsNotExist(err) {
    log.Printf("⚠️  'public' papkasi topilmadi, qidirilayotgan joy: %s", publicPath)
    log.Printf("📂 Joriy direktor: %s", currentDir)
    
    // Пробуем найти public в текущей директории
    publicPath = filepath.Join(currentDir, "public")
    if _, err := os.Stat(publicPath); os.IsNotExist(err) {
      log.Fatal("❌ 'public' papkasi hech qayerda topilmadi! Iltimos, strukturani tekshiring.")
    }
  }
  
  log.Printf("📁 Static fayllar joyi: %s", publicPath)
  
  // Выводим список файлов для отладки
  files, _ := os.ReadDir(publicPath)
  log.Printf("📄 Public papkasidagi fayllar (%d ta):", len(files))
  for _, file := range files {
    log.Printf("   - %s", file.Name())
  }

  // Настройка статического сервера
  fs := http.FileServer(http.Dir(publicPath))
  
  // Маршрут для статических файлов - должен быть ПОСЛЕДНИМ
  r.PathPrefix("/").Handler(http.StripPrefix("/", fs))

  log.Println("🚀 Server ishga tushdi")
  
  port := os.Getenv("PORT")
if port == "" {
  port = "8080"
}

log.Println("🚀 Server ishga tushdi, port:", port)
log.Fatal(http.ListenAndServe(":"+port, r))
}

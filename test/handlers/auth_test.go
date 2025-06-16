package handlers

import (
	"database/sql"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JorgeePG/todo-list/internal/handlers"
	"github.com/gorilla/sessions"

	_ "modernc.org/sqlite"
)

// App estructura que contiene la conexión a la base de datos
type App struct {
	DB *sql.DB
}

// cleanDB crea una base de datos en memoria y las tablas necesarias
func cleanDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`PRAGMA foreign_keys = ON;`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);
	CREATE TABLE tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		description TEXT,
		completed BOOLEAN,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func newTestHandler(t *testing.T) *handlers.Handler2 {
	db := cleanDB(t)
	templates := template.Must(template.ParseGlob("../../web_templates/*.html"))
	templates = template.Must(templates.ParseGlob("../../web_templates/fragments/*.html"))

	store := sessions.NewCookieStore([]byte("test-secret"))

	return &handlers.Handler2{
		Db:        db,
		Templates: templates,
		Store:     store,
	}
}

func TestRegisterHandlerGET(t *testing.T) {

	h := newTestHandler(t)

	req := httptest.NewRequest("GET", "/register", nil)
	w := httptest.NewRecorder()

	h.RegisterHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("GET /register: want %d, got %d", http.StatusOK, res.StatusCode)
	}
}

func TestRegisterHandlerOKPOST(t *testing.T) {

	form := strings.NewReader("username=testuser&password=testpass")
	req := httptest.NewRequest("POST", "/register", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h := newTestHandler(t)
	h.RegisterHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusSeeOther {
		t.Errorf("POST /register: want %d, got %d", http.StatusSeeOther, res.StatusCode)
	}
}

func TestRegisterHandlerBadPOST(t *testing.T) {
	h := newTestHandler(t)
	form := strings.NewReader("username=testuser&password=testpass")
	req1 := httptest.NewRequest("POST", "/register", form)
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w1 := httptest.NewRecorder()
	h.RegisterHandler(w1, req1)

	// Intentar registrar el mismo usuario otra vez
	req2 := httptest.NewRequest("POST", "/register", strings.NewReader("username=testuser&password=testpass"))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w2 := httptest.NewRecorder()
	h.RegisterHandler(w2, req2)

	res := w2.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("POST /register repetido: want %d, got %d", http.StatusBadRequest, res.StatusCode)
	}

	body, _ := io.ReadAll(res.Body)
	mensaje := string(body)
	if !strings.Contains(mensaje, "Usuario ya existe") {
		t.Errorf("esperaba mensaje de usuario existente, obtuve: %q", mensaje)
	}
}

package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JorgeePG/todo-list/internal/handlers"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"

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

func newTestHandler(t *testing.T) *handlers.WebHandler {
	db := cleanDB(t)
	templates := template.Must(template.ParseGlob("../../web_templates/*.html"))
	templates = template.Must(templates.ParseGlob("../../web_templates/fragments/*.html"))

	store := sessions.NewCookieStore([]byte("test-secret"))

	return &handlers.WebHandler{
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

	assert.Equal(t, http.StatusOK, res.StatusCode, "GET /register: want %d, got %d", http.StatusOK, res.StatusCode)
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

	assert.Equal(t, http.StatusSeeOther, res.StatusCode, "POST /register: want %d, got %d", http.StatusSeeOther, res.StatusCode)
}

func TestLoginHandlerOKPOST(t *testing.T) {

	form := strings.NewReader("username=testuser&password=testpass")
	req := httptest.NewRequest("POST", "/register", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h := newTestHandler(t)
	h.RegisterHandler(w, req)

	req = httptest.NewRequest("POST", "/login", strings.NewReader("username=testuser&password=testpass"))
	h.LoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusSeeOther, res.StatusCode, "POST /login: want %d, got %d", http.StatusSeeOther, res.StatusCode)
}

func TestAddTaskHandlerOKPOST(t *testing.T) {

	form := strings.NewReader("username=testuser&password=testpass")
	req := httptest.NewRequest("POST", "/register", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h := newTestHandler(t)
	h.RegisterHandler(w, req)

	req = httptest.NewRequest("POST", "/login", strings.NewReader("username=testuser&password=testpass"))
	h.LoginHandler(w, req)

	req = httptest.NewRequest("POST", "/addTask", strings.NewReader("description=Test Task"))
	w = httptest.NewRecorder()
	h.AddTask(w, req)
	res := w.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusSeeOther, res.StatusCode, "POST /addTask: want %d, got %d", http.StatusSeeOther, res.StatusCode)
}

func TestDeleteTaskHandlerOkPOST(t *testing.T) {

	form := strings.NewReader("username=testuser&password=testpass")
	req := httptest.NewRequest("POST", "/register", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h := newTestHandler(t)
	h.RegisterHandler(w, req)

	req = httptest.NewRequest("POST", "/login", strings.NewReader("username=testuser&password=testpass"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	h.LoginHandler(w, req)
	res := w.Result()
	cookies := res.Cookies()

	req = httptest.NewRequest("POST", "/addTask", strings.NewReader("description=Test Task"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w = httptest.NewRecorder()
	h.AddTask(w, req)
}

func TestDeleteTaskHandlerBadPOST(t *testing.T) {
	form := strings.NewReader("username=testuser&password=testpass")
	req := httptest.NewRequest("POST", "/register", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h := newTestHandler(t)
	h.RegisterHandler(w, req)

	// LOGIN
	req = httptest.NewRequest("POST", "/login", strings.NewReader("username=testuser&password=testpass"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	h.LoginHandler(w, req)
	res := w.Result()
	cookies := res.Cookies()

	// ADD TASK
	req = httptest.NewRequest("POST", "/addTask", strings.NewReader("description=Test Task"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w = httptest.NewRecorder()
	h.AddTask(w, req)

	// DELETE TASK (ID no existente)
	req = httptest.NewRequest("POST", "/deleteTask", strings.NewReader("task_id=9999999999"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w = httptest.NewRecorder()
	h.DeleteTask(w, req)
	res = w.Result()
	assert.Equal(t, http.StatusForbidden, res.StatusCode, "POST /deleteTask con ID no perteneciente al usuario: want %d, got %d", http.StatusBadRequest, res.StatusCode)
}

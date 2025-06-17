package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/JorgeePG/todo-list/internal/handlers"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// getTestDB crea una nueva conexión a una base de datos en memoria para cada test
func getTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// Crear tablas
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL
		);
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			done BOOLEAN,
			user_id INTEGER,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

// getTestHandler crea un nuevo WebHandler para testing
func getTestHandler(t *testing.T) *handlers.WebHandler {
	db := getTestDB(t)
	store := sessions.NewCookieStore([]byte("test-key"))
	return &handlers.WebHandler{
		Db:    db,
		Store: store,
	}
}

func TestApiRegisterHandler(t *testing.T) {
	h := getTestHandler(t)

	tests := []struct {
		name       string
		username   string
		password   string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "successful registration",
			username:   "testuser",
			password:   "testpass",
			wantStatus: http.StatusCreated,
			wantBody:   `"message":"Usuario registrado correctamente"`,
		},
		{
			name:       "duplicate username",
			username:   "testuser",
			password:   "testpass",
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"Usuario ya existe"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("username", tt.username)
			form.Add("password", tt.password)

			req := httptest.NewRequest("POST", "/api/register", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			h.ApiRegisterHandler(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}

			body := w.Body.String()
			if !strings.Contains(body, tt.wantBody) {
				t.Errorf("expected body to contain %q, got %q", tt.wantBody, body)
			}
		})
	}
}

func TestApiLoginHandler(t *testing.T) {
	h := getTestHandler(t)

	// Setup: create a test user
	hash, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	_, err := h.Db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", "testuser", hash)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		username    string
		password    string
		wantStatus  int
		wantMessage string
		wantError   string
	}{
		{
			name:        "successful login",
			username:    "testuser",
			password:    "testpass",
			wantStatus:  http.StatusOK,
			wantMessage: "Login correcto",
			wantError:   "",
		},
		{
			name:        "invalid credentials",
			username:    "testuser",
			password:    "wrongpassword",
			wantStatus:  http.StatusUnauthorized,
			wantMessage: "",
			wantError:   "Usuario o contraseña incorrectos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("username", tt.username)
			form.Add("password", tt.password)

			req := httptest.NewRequest("POST", "/api/login", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			h.ApiLoginHandler(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}

			// Parsear la respuesta JSON
			var response struct {
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			// Verificar el mensaje de éxito si se espera
			if tt.wantMessage != "" && response.Message != tt.wantMessage {
				t.Errorf("expected message %q, got %q", tt.wantMessage, response.Message)
			}

			// Verificar el mensaje de error si se espera
			if tt.wantError != "" && response.Error != tt.wantError {
				t.Errorf("expected error %q, got %q", tt.wantError, response.Error)
			}
		})
	}
}

func TestApiTaskHandlers(t *testing.T) {
	h := getTestHandler(t)

	// Setup: create a test user
	hash, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	_, err := h.Db.Exec("INSERT INTO users (id, username, password_hash) VALUES (?, ?, ?)", 1, "testuser", hash)
	if err != nil {
		t.Fatal(err)
	}

	// Test login to get session
	loginForm := url.Values{}
	loginForm.Add("username", "testuser")
	loginForm.Add("password", "testpass")
	loginReq := httptest.NewRequest("POST", "/api/login", strings.NewReader(loginForm.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginW := httptest.NewRecorder()
	h.ApiLoginHandler(loginW, loginReq)

	// Get session cookie
	cookie := loginW.Result().Cookies()[0]

	t.Run("Add Task", func(t *testing.T) {
		form := url.Values{}
		form.Add("title", "Test Task")
		form.Add("done", "true")

		req := httptest.NewRequest("POST", "/api/tasks", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(cookie)
		w := httptest.NewRecorder()

		h.ApiAddTask(w, req)

		if w.Result().StatusCode != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, w.Result().StatusCode)
		}
	})
	h = getTestHandler(t)
	t.Run("List Tasks", func(t *testing.T) {
		// Primero agregamos una tarea para listar
		_, err := h.Db.Exec("INSERT INTO tasks (title, done, user_id) VALUES (?, ?, ?)", "Test Task", false, 1)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest("GET", "/api/tasks", nil)
		req.AddCookie(cookie)
		w := httptest.NewRecorder()

		h.ApiListTasks(w, req)

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Result().StatusCode)
		}

		body := w.Body.String()
		if !strings.Contains(body, `"tasks":`) {
			t.Errorf("expected tasks in response, got %q", body)
		}
	})
	h = getTestHandler(t)
	t.Run("Update Task", func(t *testing.T) {
		// First create a task to update
		_, err := h.Db.Exec("INSERT INTO tasks (id, title, done, user_id) VALUES (?, ?, ?, ?)", 1, "Original Task", false, 1)
		if err != nil {
			t.Fatal(err)
		}

		form := url.Values{}
		form.Add("id", "1")
		form.Add("title", "Updated Task")
		form.Add("done", "true")

		req := httptest.NewRequest("PUT", "/api/tasks/1", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(cookie)
		w := httptest.NewRecorder()

		h.ApiUpdateTask(w, req)

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Result().StatusCode)
		}
	})
	h = getTestHandler(t)
	t.Run("Delete Task", func(t *testing.T) {
		// Primero crear una tarea para eliminar
		_, err := h.Db.Exec("INSERT INTO tasks (title, done, user_id) VALUES (?, ?, ?)",
			"Task to delete", false, 1)
		if err != nil {
			t.Fatal(err)
		}

		// Obtener el ID generado automáticamente
		var taskID int64
		err = h.Db.QueryRow("SELECT last_insert_rowid()").Scan(&taskID)
		if err != nil {
			t.Fatal(err)
		}

		form := url.Values{}
		form.Add("id", strconv.FormatInt(taskID, 10))

		req := httptest.NewRequest("POST", "/api/delete-task", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		// Crear el formulario con el ID

		req.AddCookie(cookie)
		w := httptest.NewRecorder()

		h.ApiDeleteTask(w, req)

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d. Response body: %s",
				http.StatusOK, w.Result().StatusCode, w.Body.String())
		}
	})
}

package handlers

import (
	"encoding/json"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/JorgeePG/todo-list/internal/models"
	"github.com/gorilla/sessions"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/crypto/bcrypt"
)

type Datos struct {
	Título string
	Texto  string
}

type PageData struct {
	Título string
	Texto  string
	Tasks  []*models.Task
	Error  string
}

type WebHandler struct {
	Db        boil.ContextExecutor
	Templates *template.Template
	Store     *sessions.CookieStore
}

func (h *WebHandler) Handler(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	dbTasks, err := models.Tasks(models.TaskWhere.UserID.EQ(null.Int64From(int64(userID)))).All(r.Context(), h.Db)
	if err != nil {
		http.Error(w, "Error obteniendo tareas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Título: "Mi To-Do List",
		Texto:  "Bienvenido a tu lista de tareas",
		Tasks:  dbTasks,
	}

	err = h.Templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, "Error ejecutando plantilla: "+err.Error(), http.StatusInternalServerError)
	}
}

type ErrorData struct {
	Error string
}

func (h *WebHandler) AddTask(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		done := r.FormValue("done")

		task := &models.Task{
			Title:  title,
			Done:   null.Bool{Bool: done == "on", Valid: true},
			ID:     generateUniqueID(),
			UserID: null.Int64From(int64(userID)),
		}
		err := task.Insert(r.Context(), h.Db, boil.Infer())
		if err != nil {
			data := ErrorData{Error: "Error insertando tarea: " + err.Error()}
			h.Templates.ExecuteTemplate(w, "addTask.html", data)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err := h.Templates.ExecuteTemplate(w, "addTask.html", nil)
	if err != nil {
		http.Error(w, "Error ejecutando plantilla: "+err.Error(), http.StatusInternalServerError)
	}
}

func generateUniqueID() null.Int64 {
	// Combinamos tiempo en nanosegundos + un número aleatorio para evitar colisiones.
	uniqueID := time.Now().UnixNano() + rand.Int63n(1000) // rand.Int63n añade entropía
	return null.Int64From(uniqueID)
}

func (h *WebHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost || r.Method == http.MethodGet {
		id := r.FormValue("id")
		var intID int64
		if id != "" {
			var err error
			intID, err = strconv.ParseInt(id, 10, 64)
			if err != nil {
				http.Error(w, "ID inválido: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		task, err := models.FindTask(r.Context(), h.Db, null.Int64From(intID))
		if err != nil || !task.UserID.Valid || task.UserID.Int64 != int64(userID) {
			http.Error(w, "No autorizado", http.StatusForbidden)
			return
		}

		_, err = task.Delete(r.Context(), h.Db)
		if err != nil {
			http.Error(w, "Error eliminando tarea: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
}

func (h *WebHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	title := r.FormValue("title")
	done := r.FormValue("done")

	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	task, err := models.FindTask(r.Context(), h.Db, null.Int64From(intID))
	if err != nil || !task.UserID.Valid || task.UserID.Int64 != int64(userID) {
		http.Error(w, "No autorizado", http.StatusForbidden)
		return
	}

	task.Title = title
	task.Done = null.Bool{Bool: done == "on", Valid: true}
	_, err = task.Update(r.Context(), h.Db, boil.Infer())
	if err != nil {
		data := ErrorData{Error: "Error actualizando tarea: " + err.Error()}
		h.Templates.ExecuteTemplate(w, "index.html", data)
		return

	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		err := h.Templates.ExecuteTemplate(w, "register.html", nil)

		if err != nil {
			http.Error(w, "Error ejecutando plantilla: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	result, err := h.Db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, hash)
	if err != nil {
		data := ErrorData{Error: "Usuario ya existe"}
		h.Templates.ExecuteTemplate(w, "register.html", data)
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Error al obtener el ID del usuario", http.StatusInternalServerError)
		return
	}

	// Crear sesión automáticamente
	session, _ := h.Store.Get(r, "session")
	session.Values["user_id"] = int(userID)
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Error guardando sesión: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *WebHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		err := h.Templates.ExecuteTemplate(w, "login.html", nil)
		if err != nil {
			http.Error(w, "Error ejecutando plantilla: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	var id int
	var hash string
	err := h.Db.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", username).Scan(&id, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		data := ErrorData{Error: "Usuario o contraseña incorrectos"}
		h.Templates.ExecuteTemplate(w, "login.html", data)
		return
	}
	// Guardar el user_id en la cookie de sesión
	session, _ := h.Store.Get(r, "session")
	session.Values["user_id"] = id
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *WebHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "session")
	delete(session.Values, "user_id")
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Manejar diferentes tipos de payload
	switch v := payload.(type) {
	case string:
		// Si es un string, asumimos que ya es JSON válido
		w.Write([]byte(v))
	default:
		// Para otros tipos, usar el encoder directamente
		json.NewEncoder(w).Encode(payload)
	}
}

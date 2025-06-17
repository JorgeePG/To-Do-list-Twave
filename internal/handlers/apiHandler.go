package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/JorgeePG/todo-list/internal/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/crypto/bcrypt"
)

func (h *WebHandler) ApiRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, `{"error":"Método no permitido"}`)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	//fmt.Printf("Registro - Password: %s, Hash: %s\n", password, string(hash))
	result, err := h.Db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, hash)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, `{"error":"Usuario ya existe"}`)
		return
	}
	userID, err := result.LastInsertId()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, `{"error":"Error al obtener el ID del usuario"}`)
		return
	}
	session, _ := h.Store.Get(r, "session")
	session.Values["user_id"] = int(userID)
	session.Save(r, w)
	writeJSON(w, http.StatusCreated, `{"message":"Usuario registrado correctamente"}`)
}

func (h *WebHandler) ApiLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Método no permitido"})
		return
	}
	username := strings.TrimSpace(r.FormValue("username"))
	password := strings.TrimSpace(r.FormValue("password"))

	var id int
	var storedHash string // Variable separada para el hash almacenado

	err := h.Db.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", username).Scan(&id, &storedHash)
	//fmt.Printf("Login - Password: %s, Stored Hash: %s\n", password, storedHash)

	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Usuario o contraseña incorrectos"})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error de base de datos"})
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Usuario o contraseña incorrectos"})
		return
	}

	session, _ := h.Store.Get(r, "session")
	session.Values["user_id"] = id
	if err := session.Save(r, w); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error guardando sesión"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Login correcto"})
}

func (h *WebHandler) ApiLogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "session")
	// Elimina todos los valores de la sesión
	for k := range session.Values {
		delete(session.Values, k)
	}
	// Invalida la cookie de sesión
	session.Options.MaxAge = -1
	session.Save(r, w)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Logout correcto"})
}

func (h *WebHandler) ApiAddTask(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "No autorizado"})
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Método no permitido"})
		return
	}
	title := r.FormValue("title")
	done := r.FormValue("done")
	task := &models.Task{
		Title:  title,
		Done:   null.Bool{Bool: done == "on" || done == "true", Valid: true},
		ID:     generateUniqueID(),
		UserID: null.Int64From(int64(userID)),
	}
	err := task.Insert(r.Context(), h.Db, boil.Infer())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error insertando tarea: " + err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"message": "Tarea creada", "task": task})
}

func (h *WebHandler) ApiDeleteTask(w http.ResponseWriter, r *http.Request) {
	// 1. Parsear el formulario primero (esto lee r.Body)
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Formulario inválido"})
		return
	}

	// 2. Ahora podemos diagnosticar y usar FormValue
	fmt.Printf("Delete Task - Form values: %v\n", r.Form)
	id := r.FormValue("id")
	fmt.Printf("Delete Task - ID: %s\n", id)

	// 3. Verificación de sesión (después del parseo)
	session, _ := h.Store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "No autorizado"})
		return
	}

	// 4. Verificación del método HTTP
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Método no permitido"})
		return
	}

	// Resto del handler...
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ID inválido"})
		return
	}

	task, err := models.FindTask(r.Context(), h.Db, null.Int64From(intID))
	if err != nil || !task.UserID.Valid || task.UserID.Int64 != int64(userID) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "No autorizado"})
		return
	}

	_, err = task.Delete(r.Context(), h.Db)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error eliminando tarea: " + err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Tarea eliminada"})
}

func (h *WebHandler) ApiUpdateTask(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "No autorizado"})
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Método no permitido"})
		return
	}
	id := r.FormValue("id")
	title := r.FormValue("title")
	done := r.FormValue("done")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ID inválido"})
		return
	}
	task, err := models.FindTask(r.Context(), h.Db, null.Int64From(intID))
	if err != nil || !task.UserID.Valid || task.UserID.Int64 != int64(userID) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "No autorizado"})
		return
	}
	task.Title = title
	task.Done = null.Bool{Bool: done == "on" || done == "true", Valid: true}
	_, err = task.Update(r.Context(), h.Db, boil.Infer())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error actualizando tarea: " + err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"message": "Tarea actualizada", "task": task})
}

func (h *WebHandler) ApiListTasks(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "No autorizado"})
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Método no permitido"})
		return
	}
	dbTasks, err := models.Tasks(models.TaskWhere.UserID.EQ(null.Int64From(int64(userID)))).All(r.Context(), h.Db)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error obteniendo tareas"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"tasks": dbTasks})
}

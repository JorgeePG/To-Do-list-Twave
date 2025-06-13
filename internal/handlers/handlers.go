package handlers

import (
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/JorgeePG/todo-list/internal/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type Datos struct {
	Título string
	Texto  string
}

type PageData struct {
	Título string
	Texto  string
	Tasks  []*models.Task // Usa el struct del modelo directamente
}

var Db boil.ContextExecutor

func Handler(w http.ResponseWriter, r *http.Request) {
	// Obtener tareas de la base de datos
	dbTasks, err := models.Tasks().All(r.Context(), Db)
	if err != nil {
		http.Error(w, "Error obteniendo tareas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Título: "Mi To-Do List",
		Texto:  "Bienvenido a tu lista de tareas",
		Tasks:  dbTasks, // Pasa directamente las tareas del modelo
	}
	plantilla, err := template.ParseFiles("../web_templates/index.html")
	if err != nil {
		http.Error(w, "Error cargando plantilla: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = plantilla.Execute(w, data)
	if err != nil {
		http.Error(w, "Error ejecutando plantilla: "+err.Error(), http.StatusInternalServerError)
	}
}

func AddTask(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		// Procesar datos del formulario
		title := r.FormValue("title")
		done := r.FormValue("done")

		// Crear y guardar la tarea usando el módulo models
		task := &models.Task{
			Title: title,
			Done:  null.Bool{Bool: done == "on", Valid: true},
			ID:    generateUniqueID(),
		}
		err := task.Insert(r.Context(), Db, boil.Infer())
		if err != nil {
			http.Error(w, "Error inserting task: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Si es GET, mostrar el formulario
	plantilla, err := template.ParseFiles("../web_templates/addTask.html")
	if err != nil {
		http.Error(w, "Error cargando plantilla: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = plantilla.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error ejecutando plantilla: "+err.Error(), http.StatusInternalServerError)
	}
}
func generateUniqueID() null.Int64 {
	// Combinamos tiempo en nanosegundos + un número aleatorio para evitar colisiones.
	uniqueID := time.Now().UnixNano() + rand.Int63n(1000) // rand.Int63n añade entropía
	return null.Int64From(uniqueID)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	// Permitir GET además de POST
	if r.Method == http.MethodPost || r.Method == http.MethodGet {
		id := r.FormValue("id") // Esto funciona tanto para POST como para GET

		// Convertir id a int64
		var intID int64
		if id != "" {
			var err error
			intID, err = strconv.ParseInt(id, 10, 64)
			if err != nil {
				http.Error(w, "ID inválido: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		// Eliminar la tarea usando el módulo models
		task := &models.Task{ID: null.Int64From(intID)}
		_, err := task.Delete(r.Context(), Db)
		if err != nil {
			http.Error(w, "Error eliminando tarea: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
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

	task, err := models.FindTask(r.Context(), Db, null.Int64From(intID))
	if err != nil {
		http.Error(w, "Tarea no encontrada: "+err.Error(), http.StatusNotFound)
		return
	}

	task.Title = title
	task.Done = null.Bool{Bool: done == "on", Valid: true}
	_, err = task.Update(r.Context(), Db, boil.Infer())
	if err != nil {
		http.Error(w, "Error actualizando tarea: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

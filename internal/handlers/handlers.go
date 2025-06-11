package handlers

import (
	"html/template"
	"math/rand"
	"net/http"
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

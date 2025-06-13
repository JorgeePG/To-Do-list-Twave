package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/JorgeePG/todo-list/internal/handlers"
	"github.com/JorgeePG/todo-list/internal/midleware"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/urfave/cli/v2"
	_ "modernc.org/sqlite"
)

var templates *template.Template

func startServer() {
	db, err := sql.Open("sqlite", "../todo.db")
	if err != nil {
		log.Fatal(err)
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,
    done BOOLEAN
	)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
	);`)

	templates = template.Must(template.ParseGlob("../web_templates/*.html"))
	templates = template.Must(templates.ParseGlob("../web_templates/fragments/*.html"))

	store := sessions.NewCookieStore([]byte("super-secret-key"))
	handlers.Store = store
	midleware.Store = store

	handlers.Db = db
	handlers.Templates = templates
	r := mux.NewRouter()
	r.Use(midleware.CspControl)

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("../web_templates/static/"))))

	// Rutas públicas
	r.HandleFunc("/register", handlers.RegisterHandler)
	r.HandleFunc("/login", handlers.LoginHandler)
	r.HandleFunc("/logout", handlers.LogoutHandler)

	// Subrouter protegido
	s := r.PathPrefix("/").Subrouter()
	s.Use(midleware.RequireLogin)

	// Rutas protegidas
	s.HandleFunc("/", handlers.Handler)
	s.HandleFunc("/addTask", handlers.AddTask).Methods("GET", "POST")
	s.HandleFunc("/delete", handlers.DeleteTask)
	s.HandleFunc("/update", handlers.UpdateTask).Methods("GET", "POST")

	log.Println("Servidor iniciado en :8080")
	http.ListenAndServe(":8080", r)
}

func main() {
	app := &cli.App{
		Name:  "todo",
		Usage: "Gestor de tareas desde CLI y Web",
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "Inicia el servidor web",
				Action: func(c *cli.Context) error {
					startServer()
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"database/sql"
	"fmt"
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

func StartServer() {
	db, err := sql.Open("sqlite", "../todo.db")
	if err != nil {
		log.Fatal(err)
	}

	// Migraciones básicas
	db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		done BOOLEAN,
		user_id INTEGER
	)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL
	)`)

	templates = template.Must(template.ParseGlob("../web_templates/*.html"))
	templates = template.Must(templates.ParseGlob("../web_templates/fragments/*.html"))

	store := sessions.NewCookieStore([]byte("super-secret-key"))
	midleware.Store = store

	r := mux.NewRouter()
	r.Use(midleware.CspControl)

	// Archivos estáticos
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("../web_templates/static/"))))

	h := &handlers.WebHandler{
		Db:        db,
		Templates: templates,
		Store:     store,
	}

	// Web: Rutas públicas
	r.HandleFunc("/register", h.RegisterHandler)
	r.HandleFunc("/login", h.LoginHandler)
	r.HandleFunc("/logout", h.LogoutHandler)

	// Web: Rutas protegidas
	web := r.PathPrefix("/").Subrouter()
	web.Use(midleware.RequireLogin)

	web.HandleFunc("/", h.Handler)
	web.HandleFunc("/addTask", h.AddTask).Methods("GET", "POST")
	web.HandleFunc("/delete", h.DeleteTask)
	web.HandleFunc("/update", h.UpdateTask).Methods("GET", "POST")

	// API: Subrouter separado
	api := r.PathPrefix("/api").Subrouter()

	apiHandler := &handlers.WebHandler{
		Db:        db,
		Templates: templates,
		Store:     store,
	}

	// Rutas API (JSON)
	api.HandleFunc("/register", apiHandler.ApiRegisterHandler).Methods("POST")
	api.HandleFunc("/login", apiHandler.ApiLoginHandler).Methods("POST")
	api.HandleFunc("/logout", apiHandler.ApiLogoutHandler).Methods("GET")
	api.HandleFunc("/tasks", apiHandler.ApiListTasks).Methods("GET")
	api.HandleFunc("/tasks", apiHandler.ApiAddTask).Methods("POST")
	api.HandleFunc("/tasks/{id:[0-9]+}", apiHandler.ApiUpdateTask).Methods("PUT")
	api.HandleFunc("/tasks/{id:[0-9]+}", apiHandler.ApiDeleteTask).Methods("DELETE")

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
					StartServer()
					return nil
				},
			},
			{
				Name:      "list",
				Usage:     "Lista todas las tareas",
				UsageText: "go run . list",
				Action: func(c *cli.Context) error {
					db, err := sql.Open("sqlite", "../todo.db")
					if err != nil {
						return err
					}
					defer db.Close()
					rows, err := db.Query("SELECT id, title, done FROM tasks")
					if err != nil {
						return err
					}
					defer rows.Close()
					for rows.Next() {
						var id int
						var title string
						var done bool
						rows.Scan(&id, &title, &done)
						status := "Pendiente"
						if done {
							status = "Hecha"
						}
						log.Printf("[%d] %s - %s\n", id, title, status)
					}
					return nil
				},
			},
			{
				Name: "add",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "title",
						Aliases: []string{"t"},
						Usage:   "Título de la tarea",
					},
					&cli.StringFlag{
						Name:  "finish",
						Usage: "Tarea acabada (true/false)",
					},
				},
				Action: func(c *cli.Context) error {
					title := c.String("title")
					finish := c.String("finish")
					fmt.Println("Título:", title, "Tarea acabada:", finish)
					return nil
				},
			},
			{
				Name: "user",
				Subcommands: []*cli.Command{
					{
						Name: "add",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "nombre",
								Aliases: []string{"t"},
								Usage:   "Nombre del usuario",
							},
							&cli.StringFlag{
								Name:  "atareado",
								Usage: "Tareas pendientes (true/false)",
							},
						},
						Action: func(c *cli.Context) error {
							nombre := c.String("nombre")
							atareado := c.String("atareado")
							fmt.Println("Nombre:", nombre, "Tareas pendientes:", atareado)
							return nil
						},
					},
					{
						Name: "delete",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "nombre",
								Aliases: []string{"t"},
								Usage:   "Nombre del usuario",
							},
						},
						Action: func(c *cli.Context) error {
							nombre := c.String("nombre")
							fmt.Println("El usuario \"", nombre, "\" ha sido eliminado")
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

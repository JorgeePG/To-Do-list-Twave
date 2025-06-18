package main

import (
	"database/sql"
	"encoding/json"
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
	var verbose bool

	app := &cli.App{
		Name:  "todo",
		Usage: "Gestor de tareas desde CLI y Web",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       "Muestra logs detallados",
				Destination: &verbose,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "Inicia el servidor web",
				Action: func(c *cli.Context) error {
					if verbose {
						log.Println("[VERBOSE] Iniciando servidor web...")
					}
					StartServer()
					return nil
				},
			},
			{
				Name:  "list",
				Usage: "Lista todas las tareas",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Formato de salida: text|json",
						Value:   "text",
					},
					&cli.BoolFlag{
						Name:  "done-only",
						Usage: "Mostrar solo tareas completadas",
					},
					&cli.BoolFlag{
						Name:  "pending-only",
						Usage: "Mostrar solo tareas pendientes",
					},
					&cli.StringFlag{
						Name:  "sort",
						Usage: "Ordenar por: id|title|status",
						Value: "id",
					},
				},
				Action: func(c *cli.Context) error {
					if verbose {
						log.Println("[VERBOSE] Conectando a la base de datos para listar tareas...")
					}

					// Construir la consulta SQL con filtros
					query := "SELECT id, title, done FROM tasks"

					// Aplicar filtros
					var whereConditions []string
					if c.Bool("done-only") {
						whereConditions = append(whereConditions, "done = 1")
					}
					if c.Bool("pending-only") {
						whereConditions = append(whereConditions, "done = 0")
					}

					// Si ambos están activados, esto sería un error lógico
					if c.Bool("done-only") && c.Bool("pending-only") {
						return fmt.Errorf("error: no puedes usar --done-only y --pending-only al mismo tiempo")
					}

					// Agregar las condiciones WHERE si existen
					if len(whereConditions) > 0 {
						query += " WHERE " + whereConditions[0]
					}

					// Aplicar orden
					sortColumn := c.String("sort")
					switch sortColumn {
					case "id":
						query += " ORDER BY id ASC"
					case "title":
						query += " ORDER BY title ASC"
					case "status":
						query += " ORDER BY done ASC"
					default:
						// Valor predeterminado en caso de valor inválido
						query += " ORDER BY id ASC"
					}

					if verbose {
						log.Printf("[VERBOSE] Ejecutando consulta: %s\n", query)
					}

					db, err := sql.Open("sqlite", "../todo.db")
					if err != nil {
						return err
					}
					defer db.Close()

					rows, err := db.Query(query)
					if err != nil {
						return err
					}
					defer rows.Close()

					type Task struct {
						ID    int    `json:"id"`
						Title string `json:"title"`
						Done  bool   `json:"done"`
					}
					var tasks []Task

					for rows.Next() {
						var t Task
						rows.Scan(&t.ID, &t.Title, &t.Done)
						tasks = append(tasks, t)
					}

					output := c.String("output")
					switch output {
					case "json":
						importjson, _ := json.MarshalIndent(tasks, "", "  ")
						fmt.Println(string(importjson))
					default:
						for _, t := range tasks {
							status := "Pendiente"
							if t.Done {
								status = "Hecha"
							}
							fmt.Printf("[%d] %s - %s\n", t.ID, t.Title, status)
						}
					}
					if verbose {
						log.Printf("[VERBOSE] %d tareas listadas.\n", len(tasks))
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

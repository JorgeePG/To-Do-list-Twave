package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/JorgeePG/todo-list/internal/handlers"
	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	_ "modernc.org/sqlite"
)

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
	handlers.Db = db
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("../web_templates/static/"))))
	r.HandleFunc("/", handlers.Handler)
	r.HandleFunc("/addTask", handlers.AddTask).Methods("GET", "POST")
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
			// Puedes añadir más comandos aquí, por ejemplo: add, list, delete, etc.
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

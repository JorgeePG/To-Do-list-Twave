package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/JorgeePG/todo-list/internal/handlers"
	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "../todo.db")
	if err != nil {
		log.Fatal(err)
	}
	handlers.Db = db
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.Handler)
	r.HandleFunc("/addTask", handlers.AddTask).Methods("GET", "POST")
	http.ListenAndServe(":8080", r)
}

package test

import (
	"testing"
)

// Ejemplo de estructura de tarea (ajusta según tu implementación)
type Task struct {
	ID    int
	Title string
	Done  bool
}

// Simulación de funciones de la aplicación (ajusta según tu código real)
func AddTask(tasks []Task, title string) []Task {
	id := len(tasks) + 1
	return append(tasks, Task{ID: id, Title: title, Done: false})
}

func CompleteTask(tasks []Task, id int) []Task {
	for i, t := range tasks {
		if t.ID == id {
			tasks[i].Done = true
		}
	}
	return tasks
}

func DeleteTask(tasks []Task, id int) []Task {
	newTasks := []Task{}
	for _, t := range tasks {
		if t.ID != id {
			newTasks = append(newTasks, t)
		}
	}
	return newTasks
}

// Test para agregar una tarea
func TestAddTask(t *testing.T) {
	tasks := []Task{}
	tasks = AddTask(tasks, "Comprar pan")
	if len(tasks) != 1 {
		t.Errorf("Se esperaba 1 tarea, se obtuvo %d", len(tasks))
	}
	if tasks[0].Title != "Comprar pan" {
		t.Errorf("El título de la tarea no coincide")
	}
}

// Test para completar una tarea
func TestCompleteTask(t *testing.T) {
	tasks := []Task{{ID: 1, Title: "Leer libro", Done: false}}
	tasks = CompleteTask(tasks, 1)
	if !tasks[0].Done {
		t.Errorf("La tarea debería estar marcada como completada")
	}
}

// Test para eliminar una tarea
func TestDeleteTask(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "Tarea 1", Done: false},
		{ID: 2, Title: "Tarea 2", Done: false},
	}
	tasks = DeleteTask(tasks, 1)
	if len(tasks) != 1 {
		t.Errorf("Se esperaba 1 tarea después de eliminar, se obtuvo %d", len(tasks))
	}
	if tasks[0].ID != 2 {
		t.Errorf("La tarea restante debería tener ID 2")
	}
}

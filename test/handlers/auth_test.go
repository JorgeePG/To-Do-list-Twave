package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JorgeePG/todo-list/internal/handlers"
)

func TestRegisterHandlerGET(t *testing.T) {
	req, err := http.NewRequest("GET", "/register", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.RegisterHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

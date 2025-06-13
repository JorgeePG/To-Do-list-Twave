package handlers

import (
	"net/http"
	"testing"
)

func TestRegisterHandlerGET(t *testing.T) {
	baseURL := "http://localhost:8080"
	req, err := http.Get(baseURL + "/register")
	if err != nil {
		t.Fatal(err)
	}
	if req.StatusCode != http.StatusOK {
		t.Errorf("GET /register: want %d, got %d", http.StatusOK, req.StatusCode)
	}

}

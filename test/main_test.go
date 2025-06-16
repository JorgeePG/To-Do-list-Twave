package test

import (
	"net/http"
	"strings"
	"testing"
)

func TestServerEndpoints(t *testing.T) {
	baseURL := "http://localhost:8080"

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		body       string
	}{
		{"GET /register", "GET", "/register", http.StatusOK, ""},
		{"GET /login", "GET", "/login", http.StatusOK, ""},
		{"POST /login invalid", "POST", "/login", http.StatusUnauthorized, "username=nouser&password=bad"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			var err error

			switch tt.method {
			case "GET":
				resp, err = http.Get(baseURL + tt.path)
			case "POST":
				resp, err = http.Post(baseURL+tt.path, "application/x-www-form-urlencoded", strings.NewReader(tt.body))
			}

			if err != nil {
				t.Fatalf("%s %s failed: %v", tt.method, tt.path, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("%s %s: want %d, got %d", tt.method, tt.path, tt.wantStatus, resp.StatusCode)
			}
		})
	}
}

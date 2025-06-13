package midleware

import "net/http"

// Middleware para agregar la cabecera CSP
func CspControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self';")
		next.ServeHTTP(w, r)
	})
}

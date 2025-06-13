package midleware

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var Store *sessions.CookieStore

func RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := Store.Get(r, "session")
		if session.Values["user_id"] == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

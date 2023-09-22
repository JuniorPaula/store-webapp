package main

import "net/http"

// Auth middleware checks whether the request contains a valid authentication
func (app *application) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := app.authenticateToken(r)
		if err != nil {
			app.errorLog.Println(err)
			app.unauthorized(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

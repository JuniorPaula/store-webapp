package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	mux.Post("/api/v1/payment-intent", app.GetPaymentIntent)

	mux.Get("/api/v1/widget/{id}", app.GetWidgetByID)

	mux.Post("/api/v1/create-customer-and-subscribe-to-plan", app.CreateCustomerAndSubscribeToPlan)

	mux.Post("/api/v1/authenticate", app.CreateAuthToken)
	mux.Post("/api/v1/is-authenticate", app.IsAuthenticated)

	mux.Route("/api/v1/admin", func(mux chi.Router) {
		mux.Use(app.Auth)

		mux.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("The secret admin test endpoint"))
		})

		mux.Post("/virtual-terminal-succeeded", app.VirtualTerminalPaymentSucceeded)
	})

	return mux
}

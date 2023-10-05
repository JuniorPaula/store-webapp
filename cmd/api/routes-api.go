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
	mux.Post("/api/v1/forgot-password", app.SendPasswordResetEmail)
	mux.Post("/api/v1/reset-password", app.ResetPassword)

	mux.Route("/api/v1/admin", func(mux chi.Router) {
		mux.Use(app.Auth)

		mux.Post("/virtual-terminal-succeeded", app.VirtualTerminalPaymentSucceeded)
		mux.Post("/all-sales", app.GetAllSales)
		mux.Post("/get-sale/{id}", app.GetSale)
		mux.Post("/all-subscriptions", app.GetAllSubscriptions)
		mux.Post("/get-subscription/{id}", app.GetSubscription)

		mux.Post("/refund", app.RefundCharge)
		mux.Post("/cancel-subscription", app.CancelSubscrition)

		mux.Post("/all-users", app.AllUsers)
		mux.Post("/all-users/{id}", app.OneUser)
	})

	return mux
}

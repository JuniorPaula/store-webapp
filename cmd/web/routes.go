package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(SessionLoad)

	mux.Get("/home", app.Home)

	mux.Get("/virtual-terminal", app.VirtualTerminal)
	mux.Post("/virtual-terminal-payment-succeeded", app.VirtualTerminalPaymentSucceeded)
	mux.Get("/virtual-terminal-receipt", app.VirtualTerminalReceipt)

	mux.Get("/receipt", app.Receipt)
	mux.Post("/payment-succeeded", app.PaymentSucceeded)
	mux.Get("/widget/{id}", app.ChargeOne)

	// Create a file server which serves files out of the "./static" directory.
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}

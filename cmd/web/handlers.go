package main

import (
	"log"
	"net/http"
	"strconv"
	"webapp/internal/card"

	"github.com/go-chi/chi/v5"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "home", &templateData{}); err != nil {
		app.errorLog.Println(err.Error())
		return
	}
}

// VirtualTerminal is a handler which renders a page where the user can enter their payment details.
func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "terminal", &templateData{}, "stripe-js"); err != nil {
		app.errorLog.Println(err.Error())
		return
	}
}

// PaymentSucceeded is a handler which renders a page to confirm that the payment succeeded.
func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err.Error())
		return
	}

	cardHolder := r.Form.Get("cardholder_name")
	email := r.Form.Get("cardholder_email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")

	card := card.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err.Error())
		return
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err.Error())
		return
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	// create a new customer

	// create a new order

	// create a new transaction

	data := make(map[string]interface{})
	data["cardholder"] = cardHolder
	data["email"] = email
	data["paymentIntent"] = paymentIntent
	data["paymentMethod"] = paymentMethod
	data["paymentAmount"] = paymentAmount
	data["paymentCurrency"] = paymentCurrency
	data["lastFour"] = lastFour
	data["expiryMonth"] = expiryMonth
	data["expiryYear"] = expiryYear
	data["bankReturnCode"] = pi.LatestCharge.ID

	// should be a redirect to a new page

	if err := app.renderTemplate(w, r, "payment-succeeded", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err.Error())
		return
	}

}

// ChargeOne is a handler which renders a page where the user can change the
// quantity of a product they want to buy.
func (app *application) ChargeOne(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	widgetID, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		log.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	err = app.renderTemplate(w, r, "buy-once", &templateData{
		Data: data,
	}, "stripe-js")

	if err != nil {
		app.errorLog.Println(err.Error())
		return
	}
}

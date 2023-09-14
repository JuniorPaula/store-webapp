package main

import "net/http"

// VirtualTerminal is a handler which renders a page where the user can enter their payment details.
func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["publishable_key"] = app.config.stripe.key

	if err := app.renderTemplate(w, r, "terminal", &templateData{StringMap: stringMap}); err != nil {
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

	data := make(map[string]interface{})
	data["cardholder"] = cardHolder
	data["email"] = email
	data["paymentIntent"] = paymentIntent
	data["paymentMethod"] = paymentMethod
	data["paymentAmount"] = paymentAmount
	data["paymentCurrency"] = paymentCurrency

	if err := app.renderTemplate(w, r, "payment-succeeded", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err.Error())
		return
	}

}

// ChargeOne is a handler which renders a page where the user can change the
// quantity of a product they want to buy.
func (app *application) ChargeOne(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "buy-once", nil); err != nil {
		app.errorLog.Println(err.Error())
		return
	}
}

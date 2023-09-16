package main

import (
	"log"
	"net/http"
	"strconv"
	"time"
	"webapp/internal/card"
	"webapp/internal/models"

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

	firtsName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	cardHolder := r.Form.Get("cardholder_name")
	email := r.Form.Get("cardholder_email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")
	widgetID, _ := strconv.Atoi(r.Form.Get("product_id"))

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
	customerID, err := app.SaveCustomer(firtsName, lastName, email)
	if err != nil {
		app.errorLog.Println("[ERROR] to save a customer:", err.Error())
		return
	}

	// create a new transaction
	amount, _ := strconv.Atoi(paymentAmount)
	txn := models.Transaction{
		Amount:              amount,
		Currency:            paymentCurrency,
		Lastfour:            lastFour,
		ExpiryMonth:         int(expiryMonth),
		ExpiryYear:          int(expiryYear),
		BankReturnCode:      pi.LatestCharge.ID,
		TransactionStatusID: 2,
	}

	txnID, err := app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println("[ERROR] to save a transaction:", err.Error())
		return
	}

	// create a new order
	order := models.Order{
		WidgetID:      widgetID,
		TransactionID: txnID,
		CustomerID:    customerID,
		StatusID:      1,
		Quantity:      1,
		Amount:        amount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_, err = app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println("[ERROR] to save an order:", err.Error())
		return
	}

	data := make(map[string]interface{})
	data["firstName"] = firtsName
	data["lastName"] = lastName
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

// SaveOrder saves an order to the database and returns id.
func (app *application) SaveOrder(order models.Order) (int, error) {
	orderID, err := app.DB.InsertOrder(order)
	if err != nil {
		return 0, err
	}

	return orderID, nil
}

// SaveTransaction saves a transaction to the database and returns id.
func (app *application) SaveTransaction(txn models.Transaction) (int, error) {
	txnID, err := app.DB.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}

	return txnID, nil
}

// SaveCustomer saves a customer to the database and returns id.
func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	customerID, err := app.DB.InsertCustomer(customer)
	if err != nil {
		return 0, err
	}

	return customerID, nil
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

package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"webapp/internal/card"
	"webapp/internal/encryption"
	"webapp/internal/models"
	"webapp/internal/urlsigner"

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
	if err := app.renderTemplate(w, r, "terminal", &templateData{}); err != nil {
		app.errorLog.Println(err.Error())
		return
	}
}

type TransactionData struct {
	FirtsName       string
	LastName        string
	Email           string
	PaymentIntentID string
	PaymentMethodID string
	PaymentAmount   int
	PaymentCurrency string
	LastFour        string
	ExpiryMonth     int
	ExpiryYear      int
	BankReturnCode  string
}

// GetTransactionData gets the transaction data from the form.
func (app *application) GetTransactionData(r *http.Request) (TransactionData, error) {
	var txnData TransactionData
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err.Error())
		return txnData, err
	}

	firtsName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	email := r.Form.Get("cardholder_email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")
	amount, _ := strconv.Atoi(paymentAmount)

	card := card.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err.Error())
		return txnData, err
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err.Error())
		return txnData, err
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	txnData = TransactionData{
		FirtsName:       firtsName,
		LastName:        lastName,
		Email:           email,
		PaymentIntentID: paymentIntent,
		PaymentMethodID: paymentMethod,
		PaymentAmount:   amount,
		PaymentCurrency: paymentCurrency,
		LastFour:        lastFour,
		ExpiryMonth:     int(expiryMonth),
		ExpiryYear:      int(expiryYear),
		BankReturnCode:  pi.LatestCharge.ID,
	}

	return txnData, nil
}

// VirtualTerminalPaymentSucceeded for the virtual terminal payment succeeded.
func (app *application) VirtualTerminalPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err.Error())
		return
	}

	// create a new transaction
	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		Lastfour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		BankReturnCode:      txnData.BankReturnCode,
		TransactionStatusID: 2,
	}

	_, err = app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println("[ERROR] to save a transaction:", err.Error())
		return
	}

	app.Session.Put(r.Context(), "receipt", txnData)
	http.Redirect(w, r, "/virtual-terminal-receipt", http.StatusSeeOther)

}

// VirtualTerminalReceipt is a handler which renders a page to confirm that the payment succeeded.
func (app *application) VirtualTerminalReceipt(w http.ResponseWriter, r *http.Request) {
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn

	app.Session.Remove(r.Context(), "receipt")

	if err := app.renderTemplate(w, r, "virtual-terminal-receipt", &templateData{Data: data}); err != nil {
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

	widgetID, _ := strconv.Atoi(r.Form.Get("product_id"))

	// get transaction data
	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err.Error())
		return
	}

	// create a new customer
	customerID, err := app.SaveCustomer(txnData.FirtsName, txnData.LastName, txnData.Email)
	if err != nil {
		app.errorLog.Println("[ERROR] to save a customer:", err.Error())
		return
	}

	// create a new transaction
	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		Lastfour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		BankReturnCode:      txnData.BankReturnCode,
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
		Amount:        txnData.PaymentAmount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_, err = app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println("[ERROR] to save an order:", err.Error())
		return
	}

	app.Session.Put(r.Context(), "receipt", txnData)
	http.Redirect(w, r, "/receipt", http.StatusSeeOther)

}

// Receipt is a handler which renders a page to confirm that the payment succeeded.
func (app *application) Receipt(w http.ResponseWriter, r *http.Request) {
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn

	app.Session.Remove(r.Context(), "receipt")

	if err := app.renderTemplate(w, r, "receipt", &templateData{Data: data}); err != nil {
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

func (app *application) PremiumPlans(w http.ResponseWriter, r *http.Request) {
	widget, err := app.DB.GetWidget(2)
	if err != nil {
		log.Println("ERROR: to get widget", err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	if err := app.renderTemplate(w, r, "premium-plans", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err.Error())
		return
	}
}

func (app *application) PremiumReceipt(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "premium-receipt", &templateData{}); err != nil {
		app.errorLog.Println(err.Error())
		return
	}
}

// Login displays the login page.
func (app *application) LoginPage(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "login", &templateData{}); err != nil {
		app.errorLog.Println(err.Error())
		return
	}
}

// PostLoginPage handles the login form submission.
func (app *application) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	app.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err.Error())
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	id, err := app.DB.Authenticate(email, password)
	if err != nil {
		app.errorLog.Println(err.Error())
		app.Session.Put(r.Context(), "flash", "Invalid credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), "user_id", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout handles the logout button.
func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	app.Session.Destroy(r.Context())
	app.Session.RenewToken(r.Context())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ForgotPassword displays the forgot password page.
func (app *application) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "forgot-password", &templateData{}); err != nil {
		app.errorLog.Println(err.Error())
		return
	}
}

// PasswordReset displays the password reset page.
func (app *application) PasswordReset(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	theURL := r.RequestURI
	testURL := fmt.Sprintf("%s%s", app.config.frontend, theURL)

	signer := urlsigner.Signer{
		Secret: []byte(app.config.secretKey),
	}

	valid := signer.VerifyToken(testURL)
	if !valid {
		app.errorLog.Println("[ERROR] invalid url - tampering detected")
		return
	}

	// make sure not expired
	expired := signer.Expired(testURL, 60)
	if expired {
		app.errorLog.Println("[ERROR] invalid url - link expired")
		return
	}

	// encrypt email
	encryptor := encryption.Encryption{
		Key: []byte(app.config.secretKey),
	}

	encryptedEmail, err := encryptor.Encrypt(email)
	if err != nil {
		app.errorLog.Println("[ERROR] encrypt email", err.Error())
		return
	}

	data := make(map[string]interface{})
	data["email"] = encryptedEmail

	if err := app.renderTemplate(w, r, "password-reset", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err.Error())
		return
	}
}

func (app *application) AllSales(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "all-sales", &templateData{}); err != nil {
		app.errorLog.Println(err.Error())
		return
	}
}

func (app *application) ShowSale(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "sale", &templateData{}); err != nil {
		app.errorLog.Println(err.Error())
		return
	}
}

func (app *application) AllSubscriptions(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "all-subscriptions", &templateData{}); err != nil {
		app.errorLog.Println(err.Error())
		return
	}
}

func (app *application) ShowSubscription(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "subscription", &templateData{}); err != nil {
		app.errorLog.Println(err.Error())
		return
	}
}

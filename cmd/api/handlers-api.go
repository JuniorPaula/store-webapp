package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"webapp/internal/card"
	"webapp/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v75"
)

type stripePayload struct {
	Currency      string `json:"currency"`
	Amount        string `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	LastFour      string `json:"last_four"`
	Plan          string `json:"plan"`
	Email         string `json:"email"`
	CardBrand     string `json:"card_brand"`
	ExpiryMonth   int    `json:"expiry_month"`
	ExpiryYear    int    `json:"expiry_year"`
	ProductID     string `json:"product_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
	ID      int    `json:"id,omitempty"`
}

func (app *application) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {
	var payload stripePayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	amount, err := strconv.Atoi(payload.Amount)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	card := card.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: payload.Currency,
	}

	okay := true

	pi, msg, err := card.Charge(payload.Currency, amount)
	if err != nil {
		okay = false
	}

	if okay {
		out, err := json.MarshalIndent(pi, "", "\t")
		if err != nil {
			app.errorLog.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	} else {
		j := jsonResponse{
			OK:      false,
			Message: msg,
			Content: "",
		}

		out, err := json.MarshalIndent(j, "", "\t")
		if err != nil {
			app.errorLog.Println(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(out)

	}

}

func (app *application) GetWidgetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	widgetID, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	out, err := json.MarshalIndent(widget, "", "\t")
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (app *application) CreateCustomerAndSubscribeToPlan(w http.ResponseWriter, r *http.Request) {
	var data stripePayload
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		app.errorLog.Println("ERROR to decoded JSON:", err)
		return
	}

	card := card.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: data.Currency,
	}

	okay := true
	var subscription *stripe.Subscription
	txnMessage := "Transação efectuada com sucesso."

	stripeCustomer, msg, err := card.CreateCustomer(data.PaymentMethod, data.Email)
	if err != nil {
		app.errorLog.Println("ERROR to create customer:", err)
		okay = false
		txnMessage = msg
	}

	if okay {
		subscription, err = card.SubscribeToPlan(stripeCustomer, data.Plan, data.Email, data.LastFour, "Visa")
		if err != nil {
			app.errorLog.Println("ERROR to subscribe to plan:", err)
			okay = false
			txnMessage = msg
		}

		app.infoLog.Println("subscriptionID:", subscription.ID)
	}

	if okay {
		productID, _ := strconv.Atoi(data.ProductID)
		customerID, err := app.SaveCustomer(data.FirstName, data.LastName, data.Email)
		if err != nil {
			app.errorLog.Println("ERROR to save customer:", err)
			return
		}

		// create a new trasaction
		amount, _ := strconv.Atoi(data.Amount)
		txn := models.Transaction{
			Amount:              amount,
			Currency:            "brl",
			Lastfour:            data.LastFour,
			ExpiryMonth:         data.ExpiryMonth,
			ExpiryYear:          data.ExpiryYear,
			TransactionStatusID: 2,
		}

		txnID, err := app.SaveTransaction(txn)
		if err != nil {
			app.errorLog.Println("ERROR to insert transaction:", err)
			return
		}

		// create a order
		order := models.Order{
			WidgetID:      productID,
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
			app.errorLog.Println("ERROR to insert order:", err)
			return
		}
	}

	resp := jsonResponse{
		OK:      okay,
		Message: txnMessage,
	}

	out, err := json.MarshalIndent(resp, "", "\t")
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
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

// SaveTransaction saves a transaction to the database and returns id.
func (app *application) SaveTransaction(txn models.Transaction) (int, error) {
	txnID, err := app.DB.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}

	return txnID, nil
}

// SaveOrder saves an order to the database and returns id.
func (app *application) SaveOrder(order models.Order) (int, error) {
	orderID, err := app.DB.InsertOrder(order)
	if err != nil {
		return 0, err
	}

	return orderID, nil
}

// CreateAuthToken creates a new auth token.
func (app *application) CreateAuthToken(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &userInput)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// get the user from the database based on the email address; send error if invalid email.
	user, err := app.DB.GetUserByEmail(userInput.Email)
	if err != nil {
		app.unauthorized(w)
		return
	}

	// validate the password; send error if invalid password.
	validPassword, err := app.passwordMatches(user.Password, userInput.Password)
	if !validPassword || err != nil {
		app.unauthorized(w)
		return
	}

	// create a new AuthToken for the user.

	// send response

	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	payload.Error = false
	payload.Message = "Seccessfully authenticated."

	_ = app.writeJSON(w, http.StatusOK, payload)
}

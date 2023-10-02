package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"webapp/internal/card"
	"webapp/internal/encryption"
	"webapp/internal/models"
	"webapp/internal/urlsigner"

	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v75"
	"golang.org/x/crypto/bcrypt"
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
			PaymentIntent:       subscription.ID,
			PaymentMethod:       data.PaymentMethod,
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
	token, err := models.GenerateToken(user.ID, 24*time.Hour, models.ScopeAuthentication)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// save to the database.
	err = app.DB.InsertToken(token, user)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// send response
	var payload struct {
		Error   bool          `json:"error"`
		Message string        `json:"message"`
		Token   *models.Token `json:"authentication_token"`
	}

	payload.Error = false
	payload.Message = fmt.Sprintf("token for %s created", userInput.Email)
	payload.Token = token

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) authenticateToken(r *http.Request) (*models.User, error) {
	// get the token from the request header.
	// validate the token and get associated user.
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return nil, errors.New("authorization header required")
	}

	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, errors.New("authorization header format must be Bearer {token}")
	}

	token := headerParts[1]
	if len(token) != 26 {
		return nil, errors.New("token must be 26 characters long")
	}

	user, err := app.DB.GetUserByToken(token)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (app *application) IsAuthenticated(w http.ResponseWriter, r *http.Request) {
	// validate the token and get associated user.
	user, err := app.authenticateToken(r)
	if err != nil {
		app.unauthorized(w)
		return
	}

	// send response
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	payload.Error = false
	payload.Message = fmt.Sprintf("user %s is authenticated", user.Email)

	app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) VirtualTerminalPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	var txnData struct {
		PaymentAmount   int    `json:"amount"`
		PaymentCurrency string `json:"currency"`
		PaymentIntent   string `json:"payment_intent"`
		PaymentMethod   string `json:"payment_method"`
		BankReturnCode  string `json:"bank_return_code"`
		ExpiryMonth     int    `json:"expiry_month"`
		ExpiryYear      int    `json:"expiry_year"`
		LastFour        string `json:"last_four"`
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		Email           string `json:"email"`
	}

	err := app.readJSON(w, r, &txnData)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	card := card.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	pi, err := card.RetrievePaymentIntent(txnData.PaymentIntent)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	pm, err := card.GetPaymentMethod(txnData.PaymentMethod)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	txnData.LastFour = pm.Card.Last4
	txnData.ExpiryMonth = int(pm.Card.ExpMonth)
	txnData.ExpiryYear = int(pm.Card.ExpYear)

	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		Lastfour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		PaymentIntent:       txnData.PaymentIntent,
		PaymentMethod:       txnData.PaymentMethod,
		BankReturnCode:      pi.LatestCharge.ID,
		TransactionStatusID: 2,
	}

	_, err = app.SaveTransaction(txn)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, txn)
}

func (app *application) SendPasswordResetEmail(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// verify if the email exists in the database
	_, err = app.DB.GetUserByEmail(payload.Email)
	if err != nil {
		var resp struct {
			Error   bool   `json:"error"`
			Message string `json:"message"`
		}
		resp.Error = true
		resp.Message = "email não encontrado"

		app.writeJSON(w, http.StatusNotFound, resp)
		return
	}

	link := fmt.Sprintf("%s/password-reset?email=%s", app.config.frontend, payload.Email)

	sign := urlsigner.Signer{
		Secret: []byte(app.config.secretKey),
	}

	signedLink := sign.GenerateTokenFromString(link)

	var data struct {
		Link string
	}

	data.Link = signedLink

	err = app.SendEmail("info@widgets.com", payload.Email, "Recuperar senha", "password-reset", data)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "email enviado com sucesso"

	app.writeJSON(w, http.StatusCreated, resp)
}

func (app *application) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// decrypt the email
	encryptor := encryption.Encryption{
		Key: []byte(app.config.secretKey),
	}

	email, err := encryptor.Decrypt(payload.Email)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	user, err := app.DB.GetUserByEmail(email)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	nexHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	err = app.DB.UpdatePassword(string(nexHash), user)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	resp.Error = false
	resp.Message = "senha alterada com sucesso"

	app.writeJSON(w, http.StatusOK, resp)
}

func (app *application) GetAllSales(w http.ResponseWriter, r *http.Request) {
	allSales, err := app.DB.GetAllOrders()
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, allSales)
}

func (app *application) GetSale(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	saleID, _ := strconv.Atoi(id)

	sale, err := app.DB.GetOrderByID(saleID)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, sale)
}

func (app *application) GetAllSubscriptions(w http.ResponseWriter, r *http.Request) {
	allSubscriptions, err := app.DB.GetAllSubscriptions()
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, allSubscriptions)
}

func (app *application) GetSubscription(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	subscriptionID, _ := strconv.Atoi(id)

	subscription, err := app.DB.GetSubscriptionByID(subscriptionID)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, subscription)
}

func (app *application) RefundCharge(w http.ResponseWriter, r *http.Request) {
	var chargeToRefund struct {
		ID            int    `json:"id"`
		PaymentIntent string `json:"payment_intent"`
		Amount        int    `json:"amount"`
		Currency      string `json:"currency"`
	}

	err := app.readJSON(w, r, &chargeToRefund)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	card := card.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: chargeToRefund.Currency,
	}

	err = card.Refund(chargeToRefund.PaymentIntent, chargeToRefund.Amount)
	if err != nil {
		app.badRequest(w, r, errors.New("erro ao estornar a compra"))
		return
	}

	// update status in the database
	err = app.DB.UpdateOrderStatus(chargeToRefund.ID, 2)
	if err != nil {
		app.badRequest(w, r, errors.New("erro ao atualizar o status do pedido"))
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	resp.Error = false
	resp.Message = "cobrança estornada com sucesso"

	app.writeJSON(w, http.StatusOK, resp)
}

// CancelSubscrition cancels a subscription in stripe.
func (app *application) CancelSubscrition(w http.ResponseWriter, r *http.Request) {
	var subToCancel struct {
		ID            int    `json:"id"`
		PaymentIntent string `json:"payment_intent"`
		Currency      string `json:"currency"`
	}

	err := app.readJSON(w, r, &subToCancel)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	card := card.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: subToCancel.Currency,
	}

	err = card.CancelSubscription(subToCancel.PaymentIntent)
	if err != nil {
		app.badRequest(w, r, errors.New("erro ao cancelar a assinatura"))
		return
	}

	// update status in the database
	err = app.DB.UpdateOrderStatus(subToCancel.ID, 3)
	if err != nil {
		app.badRequest(w, r, errors.New("erro ao atualizar o status do pedido"))
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	resp.Error = false
	resp.Message = "Inscrição cancelada com sucesso"

	app.writeJSON(w, http.StatusOK, resp)
}

package card

import (
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/customer"
	"github.com/stripe/stripe-go/v75/paymentintent"
	"github.com/stripe/stripe-go/v75/paymentmethod"
	"github.com/stripe/stripe-go/v75/subscription"
)

// Card represents a card in the database.
type Card struct {
	Secret   string
	Key      string
	Currency string
}

// Transaction represents a transaction in the database.
type Transaction struct {
	TransactionStatusID int
	Amount              int
	Currency            string
	LastFour            string
	BankReturnCode      string
}

// Charge creates a payment intent in stripe.
func (c *Card) Charge(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	return c.CreatePaymentIntent(currency, amount)
}

// CreatePaymentIntent creates a payment intent in stripe.
func (c *Card) CreatePaymentIntent(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	stripe.Key = c.Secret

	// create payment intent
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount)),
		Currency: stripe.String(currency),
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		msg := ""
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = cardErrorMessage(stripeErr.Code)
		}
		return nil, msg, err
	}

	return pi, "", nil
}

// GetPaymentMethod gets a payment method from stripe.
func (c *Card) GetPaymentMethod(s string) (*stripe.PaymentMethod, error) {
	stripe.Key = c.Secret

	// get payment method from stripe
	pm, err := paymentmethod.Get(s, nil)
	if err != nil {
		return nil, err
	}

	return pm, nil
}

// RetrievePaymentIntent gets an exists payment intent from stripe.
func (c *Card) RetrievePaymentIntent(s string) (*stripe.PaymentIntent, error) {
	stripe.Key = c.Secret

	// get payment intent from stripe
	pi, err := paymentintent.Get(s, nil)
	if err != nil {
		return nil, err
	}

	return pi, nil
}

// SubscribeToPlan creates a subscription to a plan in stripe.
func (c *Card) SubscribeToPlan(customer *stripe.Customer, plan, email, last4, cardType string) (*stripe.Subscription, error) {
	stripeCustomerID := customer.ID
	items := []*stripe.SubscriptionItemsParams{
		{Plan: stripe.String(plan)},
	}
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(stripeCustomerID),
		Items:    items,
	}
	params.AddMetadata("last_four", last4)
	params.AddMetadata("card_type", cardType)
	params.AddExpand("latest_invoice.payment_intent")

	sub, err := subscription.New(params)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

// CreateCustomerAndSubscribeToPlan creates a customer and subscribes to a plan in stripe.
func (c *Card) CreateCustomer(pm, email string) (*stripe.Customer, string, error) {
	stripe.Key = c.Secret
	customerParams := &stripe.CustomerParams{
		PaymentMethod: stripe.String(pm),
		Email:         stripe.String(email),
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm),
		},
	}
	customer, err := customer.New(customerParams)
	if err != nil {
		msg := ""
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = cardErrorMessage(stripeErr.Code)
		}
		return nil, msg, err
	}
	return customer, "", nil
}

// cardErrorMessage returns a string with a description of the error code.
func cardErrorMessage(code stripe.ErrorCode) string {
	var msg = ""
	switch code {
	case stripe.ErrorCodeCardDeclined:
		msg = "Seu cartão foi recusado."

	case stripe.ErrorCodeExpiredCard:
		msg = "Seu cartão está expirado."

	case stripe.ErrorCodeIncorrectCVC:
		msg = "O código de segurança do seu cartão está incorreto."

	case stripe.ErrorCodeIncorrectZip:
		msg = "O código postal falhou na validação."

	case stripe.ErrorCodeAmountTooLarge:
		msg = "O valor do pagamento é muito grande."

	case stripe.ErrorCodeAmountTooSmall:
		msg = "O valor do pagamento é muito pequeno."

	case stripe.ErrorCodeBalanceInsufficient:
		msg = "Seu cartão não tem saldo suficiente."

	case stripe.ErrorCodePostalCodeInvalid:
		msg = "O código postal falhou na validação."
	default:
		msg = "Ocorreu um erro ao processar seu pagamento."
	}
	return msg
}

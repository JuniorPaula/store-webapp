package card

import (
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/paymentintent"
)

type Card struct {
	Secret   string
	Key      string
	Currency string
}

type Transaction struct {
	TransactionStatusID int
	Amount              int
	Currency            string
	LastFour            string
	BankReturnCode      string
}

func (c *Card) Charge(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	return c.CreatePaymentIntent(currency, amount)
}

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
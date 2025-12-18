package payment

import (
	"os"
	"strings"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
)

// StripeProvider is a simple Stripe adapter. It expects STRIPE_API_KEY to be set.
type StripeProvider struct {
	apiKey string
}

// NewStripeProviderFromEnv returns a StripeProvider if STRIPE_API_KEY is present, otherwise nil.
func NewStripeProviderFromEnv() PaymentProvider {
	key := os.Getenv("STRIPE_API_KEY")
	if strings.TrimSpace(key) == "" {
		return nil
	}
	return &StripeProvider{apiKey: key}
}

// Charge implements PaymentProvider using Stripe PaymentIntents.
// amount is in major units (e.g. 12.34) — we convert to cents.
func (s *StripeProvider) Charge(amount float64, currency string, token string, metadata map[string]string) (string, error) {
	stripe.Key = s.apiKey
	cents := int64(amount * 100)
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(cents),
		Currency: stripe.String(strings.ToLower(currency)),
		Confirm:  stripe.Bool(true),
	}
	// If frontend gave a payment method token, set it.
	if token != "" {
		params.PaymentMethod = stripe.String(token)
	}
	// attach metadata
	if metadata != nil {
		for k, v := range metadata {
			params.AddMetadata(k, v)
		}
	}
	pi, err := paymentintent.New(params)
	if err != nil {
		return "", err
	}
	return pi.ID, nil
}

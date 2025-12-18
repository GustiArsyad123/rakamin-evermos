package payment

import (
	"os"
)

// PaymentProvider defines a minimal interface for charging payments.
type PaymentProvider interface {
	// Charge charges the given amount (in major units, e.g. 12.34) with currency and token.
	// Returns provider transaction id on success.
	Charge(amount float64, currency string, token string, metadata map[string]string) (string, error)
}

// NewProviderFromEnv returns a provider implementation selected by env vars.
// - If PAYMENT_PROVIDER=stripe and STRIPE_API_KEY is set, it returns a StripeProvider.
// - Otherwise it falls back to the in-memory MockProvider for development.
func NewProviderFromEnv() PaymentProvider {
	prov := os.Getenv("PAYMENT_PROVIDER")
	switch prov {
	case "stripe":
		// NewStripeProviderFromEnv returns nil if STRIPE_API_KEY is not present.
		if sp := NewStripeProviderFromEnv(); sp != nil {
			return sp
		}
		return &MockProvider{}
	default:
		return &MockProvider{}
	}
}

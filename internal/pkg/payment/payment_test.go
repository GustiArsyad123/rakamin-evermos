package payment

import (
	"os"
	"testing"
)

func TestNewProviderFromEnv_DefaultsToMock(t *testing.T) {
	os.Unsetenv("PAYMENT_PROVIDER")
	p := NewProviderFromEnv()
	if p == nil {
		t.Fatalf("expected provider, got nil")
	}
}

func TestNewProviderFromEnv_StripeFactory(t *testing.T) {
	os.Setenv("PAYMENT_PROVIDER", "stripe")
	os.Unsetenv("STRIPE_API_KEY")
	p := NewProviderFromEnv()
	if p == nil {
		t.Fatalf("expected provider, got nil")
	}
	os.Setenv("STRIPE_API_KEY", "sk_test_dummy")
	p2 := NewProviderFromEnv()
	if p2 == nil {
		t.Fatalf("expected stripe provider when STRIPE_API_KEY set")
	}
}

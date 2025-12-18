package payment

import (
	"fmt"
)

// MockProvider is a simple in-memory stub that simulates successful charges.
type MockProvider struct{}

func (m *MockProvider) Charge(amount float64, currency string, token string, metadata map[string]string) (string, error) {
	// This mock accepts any token and returns a fake transaction id.
	// Replace with real provider integration (Stripe, Midtrans, etc.) in production.
	return fmt.Sprintf("mock_txn_%d", int64(amount*100)), nil
}

package address

import (
	"errors"
	"testing"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

// mockRepo implements minimal Repository for address tests.
type mockAddressRepo struct {
	address *models.Address
	err     error
}

func (m *mockAddressRepo) Create(a *models.Address) (int64, error) { return 0, nil }
func (m *mockAddressRepo) ListByUser(userID int64, filters map[string]string, page, limit int) ([]*models.Address, int, error) {
	return nil, 0, nil
}
func (m *mockAddressRepo) GetByID(id int64) (*models.Address, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.address, nil
}
func (m *mockAddressRepo) Update(id int64, label, address, city, postalCode string) error { return nil }
func (m *mockAddressRepo) Delete(id int64) error                                          { return nil }

func TestGetAddress_Authorization(t *testing.T) {
	repo := &mockAddressRepo{}
	u := &addressUsecase{repo: repo}

	// 1) Owner gets own address -> allowed
	repo.address = &models.Address{ID: 1, UserID: 10}
	a, err := u.GetAddress(10, 1, "user")
	if err != nil {
		t.Fatalf("expected owner get allowed, got err: %v", err)
	}
	if a == nil || a.ID != 1 {
		t.Fatalf("expected address returned")
	}

	// 2) Non-owner non-admin gets another user's address -> forbidden
	_, err = u.GetAddress(11, 1, "user")
	if err == nil || err.Error() != "forbidden" {
		t.Fatalf("expected forbidden, got err: %v", err)
	}

	// 3) Admin gets any address -> allowed
	a, err = u.GetAddress(1, 1, "admin")
	if err != nil {
		t.Fatalf("expected admin get allowed, got err: %v", err)
	}
	if a == nil {
		t.Fatalf("expected address returned")
	}

	// 4) Address not found
	repo.address = nil
	_, err = u.GetAddress(10, 1, "user")
	if err != nil {
		t.Fatalf("expected nil for not found, got err: %v", err)
	}

	// 5) Repo error -> forwarded
	repo.err = errors.New("db fail")
	_, err = u.GetAddress(10, 1, "user")
	if err == nil {
		t.Fatalf("expected repo error forwarded, got nil")
	}
}

func TestUpdateAddress_Authorization(t *testing.T) {
	repo := &mockAddressRepo{}
	u := &addressUsecase{repo: repo}

	// 1) Owner updates own address -> allowed
	repo.address = &models.Address{ID: 1, UserID: 10}
	err := u.UpdateAddress(10, 1, "user", "home", "addr", "city", "123")
	if err != nil {
		t.Fatalf("expected owner update allowed, got err: %v", err)
	}

	// 2) Non-owner non-admin updates another user's address -> forbidden
	err = u.UpdateAddress(11, 1, "user", "home", "addr", "city", "123")
	if err == nil || err.Error() != "forbidden" {
		t.Fatalf("expected forbidden, got err: %v", err)
	}

	// 3) Admin updates any address -> allowed
	err = u.UpdateAddress(1, 1, "admin", "home", "addr", "city", "123")
	if err != nil {
		t.Fatalf("expected admin update allowed, got err: %v", err)
	}

	// 4) Address not found
	repo.address = nil
	err = u.UpdateAddress(10, 1, "user", "home", "addr", "city", "123")
	if err == nil || err.Error() != "not found" {
		t.Fatalf("expected not found, got err: %v", err)
	}

	// 5) Repo error -> forwarded
	repo.err = errors.New("db fail")
	err = u.UpdateAddress(10, 1, "user", "home", "addr", "city", "123")
	if err == nil {
		t.Fatalf("expected repo error forwarded, got nil")
	}
}

func TestDeleteAddress_Authorization(t *testing.T) {
	repo := &mockAddressRepo{}
	u := &addressUsecase{repo: repo}

	// 1) Owner deletes own address -> allowed
	repo.address = &models.Address{ID: 1, UserID: 10}
	err := u.DeleteAddress(10, 1, "user")
	if err != nil {
		t.Fatalf("expected owner delete allowed, got err: %v", err)
	}

	// 2) Non-owner non-admin deletes another user's address -> forbidden
	err = u.DeleteAddress(11, 1, "user")
	if err == nil || err.Error() != "forbidden" {
		t.Fatalf("expected forbidden, got err: %v", err)
	}

	// 3) Admin deletes any address -> allowed
	err = u.DeleteAddress(1, 1, "admin")
	if err != nil {
		t.Fatalf("expected admin delete allowed, got err: %v", err)
	}

	// 4) Address not found
	repo.address = nil
	err = u.DeleteAddress(10, 1, "user")
	if err == nil || err.Error() != "not found" {
		t.Fatalf("expected not found, got err: %v", err)
	}

	// 5) Repo error -> forwarded
	repo.err = errors.New("db fail")
	err = u.DeleteAddress(10, 1, "user")
	if err == nil {
		t.Fatalf("expected repo error forwarded, got nil")
	}
}

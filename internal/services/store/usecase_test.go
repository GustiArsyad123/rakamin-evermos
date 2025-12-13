package store

import (
	"errors"
	"testing"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

// mockRepo implements minimal Repository for store tests.
type mockStoreRepo struct {
	store *models.Store
	err   error
}

func (m *mockStoreRepo) Create(s *models.Store) (int64, error) { return 0, nil }
func (m *mockStoreRepo) GetByID(id int64) (*models.Store, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.store, nil
}
func (m *mockStoreRepo) GetByUserID(userID int64) (*models.Store, error) { return nil, nil }
func (m *mockStoreRepo) Update(id int64, name string) error              { return nil }
func (m *mockStoreRepo) Delete(id int64) error                           { return nil }

func TestGetStore_Authorization(t *testing.T) {
	repo := &mockStoreRepo{}
	u := &storeUsecase{repo: repo}

	// 1) Owner gets own store -> allowed
	repo.store = &models.Store{ID: 1, UserID: 10}
	s, err := u.GetStore(10, 1, "user")
	if err != nil {
		t.Fatalf("expected owner get allowed, got err: %v", err)
	}
	if s == nil || s.ID != 1 {
		t.Fatalf("expected store returned")
	}

	// 2) Non-owner non-admin gets another user's store -> forbidden
	_, err = u.GetStore(11, 1, "user")
	if err == nil || err.Error() != "forbidden" {
		t.Fatalf("expected forbidden, got err: %v", err)
	}

	// 3) Admin gets any store -> allowed
	s, err = u.GetStore(1, 1, "admin")
	if err != nil {
		t.Fatalf("expected admin get allowed, got err: %v", err)
	}
	if s == nil {
		t.Fatalf("expected store returned")
	}

	// 4) Store not found
	repo.store = nil
	s, err = u.GetStore(10, 1, "user")
	if err != nil {
		t.Fatalf("expected nil for not found, got err: %v", err)
	}
	if s != nil {
		t.Fatalf("expected nil store for not found")
	}

	// 5) Repo error -> forwarded
	repo.err = errors.New("db fail")
	_, err = u.GetStore(10, 1, "user")
	if err == nil {
		t.Fatalf("expected repo error forwarded, got nil")
	}
}

func TestUpdateStore_Authorization(t *testing.T) {
	repo := &mockStoreRepo{}
	u := &storeUsecase{repo: repo}

	// 1) Owner updates own store -> allowed
	repo.store = &models.Store{ID: 1, UserID: 10}
	err := u.UpdateStore(10, 1, "user", "new name")
	if err != nil {
		t.Fatalf("expected owner update allowed, got err: %v", err)
	}

	// 2) Non-owner non-admin updates another user's store -> forbidden
	err = u.UpdateStore(11, 1, "user", "new name")
	if err == nil || err.Error() != "forbidden" {
		t.Fatalf("expected forbidden, got err: %v", err)
	}

	// 3) Admin updates any store -> allowed
	err = u.UpdateStore(1, 1, "admin", "new name")
	if err != nil {
		t.Fatalf("expected admin update allowed, got err: %v", err)
	}

	// 4) Store not found
	repo.store = nil
	err = u.UpdateStore(10, 1, "user", "new name")
	if err == nil || err.Error() != "not found" {
		t.Fatalf("expected not found, got err: %v", err)
	}

	// 5) Repo error -> forwarded
	repo.err = errors.New("db fail")
	err = u.UpdateStore(10, 1, "user", "new name")
	if err == nil {
		t.Fatalf("expected repo error forwarded, got nil")
	}
}

func TestDeleteStore_Authorization(t *testing.T) {
	repo := &mockStoreRepo{}
	u := &storeUsecase{repo: repo}

	// 1) Owner deletes own store -> allowed
	repo.store = &models.Store{ID: 1, UserID: 10}
	err := u.DeleteStore(10, 1, "user")
	if err != nil {
		t.Fatalf("expected owner delete allowed, got err: %v", err)
	}

	// 2) Non-owner non-admin deletes another user's store -> forbidden
	err = u.DeleteStore(11, 1, "user")
	if err == nil || err.Error() != "forbidden" {
		t.Fatalf("expected forbidden, got err: %v", err)
	}

	// 3) Admin deletes any store -> allowed
	err = u.DeleteStore(1, 1, "admin")
	if err != nil {
		t.Fatalf("expected admin delete allowed, got err: %v", err)
	}

	// 4) Store not found
	repo.store = nil
	err = u.DeleteStore(10, 1, "user")
	if err == nil || err.Error() != "not found" {
		t.Fatalf("expected not found, got err: %v", err)
	}

	// 5) Repo error -> forwarded
	repo.err = errors.New("db fail")
	err = u.DeleteStore(10, 1, "user")
	if err == nil {
		t.Fatalf("expected repo error forwarded, got nil")
	}
}

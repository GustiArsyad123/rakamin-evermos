package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

// mockRepo implements minimal Repository for UpdateUser tests.
type mockRepo struct {
	lastID    int64
	lastName  *string
	lastPhone *string
	lastRole  *string
	err       error
}

func (m *mockRepo) CreateUser(user *models.User) (int64, error)       { return 0, nil }
func (m *mockRepo) CreateStore(userID int64, name string) error       { return nil }
func (m *mockRepo) GetUserByEmail(email string) (*models.User, error) { return nil, nil }
func (m *mockRepo) GetUserByPhone(phone string) (*models.User, error) { return nil, nil }
func (m *mockRepo) UpdatePassword(userID int64, hashed string) error  { return nil }
func (m *mockRepo) GetUserByID(id int64) (*models.User, error)        { return nil, nil }
func (m *mockRepo) ListUsers(page, limit int, search string) ([]*models.User, int, error) {
	return nil, 0, nil
}
func (m *mockRepo) CreateRefreshToken(userID int64, tokenHash string, expiresAt time.Time) error {
	return nil
}
func (m *mockRepo) GetRefreshToken(tokenHash string) (int64, time.Time, bool, error) {
	return 0, time.Time{}, false, nil
}
func (m *mockRepo) DeleteRefreshToken(tokenHash string) error    { return nil }
func (m *mockRepo) DeleteRefreshTokensByUser(userID int64) error { return nil }

func (m *mockRepo) UpdateUser(id int64, name, phone, role *string) error {
	if m.err != nil {
		return m.err
	}
	m.lastID = id
	m.lastName = name
	m.lastPhone = phone
	m.lastRole = role
	return nil
}

func TestUpdateUser_Authorization(t *testing.T) {
	repo := &mockRepo{}
	u := &authUsecase{repo: repo}

	// 1) Owner updates own name -> allowed
	name := "Owner New"
	if err := u.UpdateUser(10, 10, "user", &name, nil, nil); err != nil {
		t.Fatalf("expected owner update allowed, got err: %v", err)
	}
	if repo.lastID != 10 || repo.lastName == nil || *repo.lastName != name {
		t.Fatalf("unexpected repo update values: %#v", repo)
	}

	// 2) Non-owner non-admin updating another user -> forbidden
	if err := u.UpdateUser(11, 12, "user", &name, nil, nil); err == nil {
		t.Fatalf("expected forbidden, got nil")
	}

	// 3) Non-admin attempting to change role -> forbidden
	role := "admin"
	if err := u.UpdateUser(10, 10, "user", nil, nil, &role); err == nil {
		t.Fatalf("expected forbidden for role change by non-admin, got nil")
	}

	// 4) Admin changing role of another user -> allowed
	if err := u.UpdateUser(1, 20, "admin", nil, nil, &role); err != nil {
		t.Fatalf("expected admin role change allowed, got err: %v", err)
	}
	if repo.lastID != 20 || repo.lastRole == nil || *repo.lastRole != role {
		t.Fatalf("unexpected repo update values for role change: %#v", repo)
	}

	// 5) Repo returns error -> forwarded
	repo.err = errors.New("db fail")
	if err := u.UpdateUser(1, 20, "admin", nil, nil, &role); err == nil {
		t.Fatalf("expected repo error forwarded, got nil")
	}
}

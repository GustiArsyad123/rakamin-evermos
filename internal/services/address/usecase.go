package address

import (
	"errors"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

type Usecase interface {
	CreateAddress(userID int64, a *models.Address) (int64, error)
	ListAddresses(userID int64, filters map[string]string, page, limit int) ([]*models.Address, int, error)
	GetAddress(userID, id int64) (*models.Address, error)
	UpdateAddress(userID, id int64, label, address, city, postalCode string) error
	DeleteAddress(userID, id int64) error
}

type addressUsecase struct {
	repo Repository
}

func NewUsecase(r Repository) Usecase {
	return &addressUsecase{repo: r}
}

func (u *addressUsecase) CreateAddress(userID int64, a *models.Address) (int64, error) {
	a.UserID = userID
	return u.repo.Create(a)
}

func (u *addressUsecase) ListAddresses(userID int64, filters map[string]string, page, limit int) ([]*models.Address, int, error) {
	return u.repo.ListByUser(userID, filters, page, limit)
}

func (u *addressUsecase) GetAddress(userID, id int64) (*models.Address, error) {
	a, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, nil
	}
	if a.UserID != userID {
		return nil, errors.New("forbidden")
	}
	return a, nil
}

func (u *addressUsecase) UpdateAddress(userID, id int64, label, address, city, postalCode string) error {
	// Check ownership
	a, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("not found")
	}
	if a.UserID != userID {
		return errors.New("forbidden")
	}
	return u.repo.Update(id, label, address, city, postalCode)
}

func (u *addressUsecase) DeleteAddress(userID, id int64) error {
	// Check ownership
	a, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("not found")
	}
	if a.UserID != userID {
		return errors.New("forbidden")
	}
	return u.repo.Delete(id)
}

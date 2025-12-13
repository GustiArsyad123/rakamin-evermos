package address

import (
	"errors"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

type Usecase interface {
	CreateAddress(userID int64, a *models.Address) (int64, error)
	ListAddresses(userID int64, filters map[string]string, page, limit int) ([]*models.Address, int, error)
	GetAddress(requesterID, id int64, requesterRole string) (*models.Address, error)
	UpdateAddress(requesterID, id int64, requesterRole, label, address, city, postalCode string) error
	DeleteAddress(requesterID, id int64, requesterRole string) error
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

func (u *addressUsecase) GetAddress(requesterID, id int64, requesterRole string) (*models.Address, error) {
	a, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, nil
	}
	if requesterRole != "admin" && a.UserID != requesterID {
		return nil, errors.New("forbidden")
	}
	return a, nil
}

func (u *addressUsecase) UpdateAddress(requesterID, id int64, requesterRole, label, address, city, postalCode string) error {
	// Check ownership
	a, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("not found")
	}
	if requesterRole != "admin" && a.UserID != requesterID {
		return errors.New("forbidden")
	}
	return u.repo.Update(id, label, address, city, postalCode)
}

func (u *addressUsecase) DeleteAddress(requesterID, id int64, requesterRole string) error {
	// Check ownership
	a, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("not found")
	}
	if requesterRole != "admin" && a.UserID != requesterID {
		return errors.New("forbidden")
	}
	return u.repo.Delete(id)
}

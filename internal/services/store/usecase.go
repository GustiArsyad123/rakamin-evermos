package store

import (
	"errors"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

type Usecase interface {
	CreateStore(userID int64, name string) (int64, error)
	GetStore(requesterID, storeID int64, requesterRole string) (*models.Store, error)
	UpdateStore(requesterID, storeID int64, requesterRole, name string) error
	DeleteStore(requesterID, storeID int64, requesterRole string) error
}

type storeUsecase struct {
	repo Repository
}

func NewUsecase(r Repository) Usecase {
	return &storeUsecase{repo: r}
}

func (u *storeUsecase) CreateStore(userID int64, name string) (int64, error) {
	// Check if user already has a store
	_, err := u.repo.GetByUserID(userID)
	if err == nil {
		return 0, errors.New("user already has a store")
	}
	if err != nil && err.Error() != "sql: no rows in result set" {
		return 0, err
	}
	s := &models.Store{UserID: userID, Name: name}
	return u.repo.Create(s)
}

func (u *storeUsecase) GetStore(requesterID, storeID int64, requesterRole string) (*models.Store, error) {
	s, err := u.repo.GetByID(storeID)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, nil
	}
	if requesterRole != "admin" && s.UserID != requesterID {
		return nil, errors.New("forbidden")
	}
	return s, nil
}

func (u *storeUsecase) UpdateStore(requesterID, storeID int64, requesterRole, name string) error {
	s, err := u.repo.GetByID(storeID)
	if err != nil {
		return err
	}
	if s == nil {
		return errors.New("not found")
	}
	if requesterRole != "admin" && s.UserID != requesterID {
		return errors.New("forbidden")
	}
	return u.repo.Update(storeID, name)
}

func (u *storeUsecase) DeleteStore(requesterID, storeID int64, requesterRole string) error {
	s, err := u.repo.GetByID(storeID)
	if err != nil {
		return err
	}
	if s == nil {
		return errors.New("not found")
	}
	if requesterRole != "admin" && s.UserID != requesterID {
		return errors.New("forbidden")
	}
	return u.repo.Delete(storeID)
}

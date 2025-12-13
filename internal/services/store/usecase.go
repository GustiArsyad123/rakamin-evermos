package store

import (
	"errors"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

type Usecase interface {
	CreateStore(userID int64, name string) (int64, error)
	GetStore(userID, storeID int64) (*models.Store, error)
	UpdateStore(userID, storeID int64, name string) error
	DeleteStore(userID, storeID int64) error
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

func (u *storeUsecase) GetStore(userID, storeID int64) (*models.Store, error) {
	s, err := u.repo.GetByID(storeID)
	if err != nil {
		return nil, err
	}
	if s.UserID != userID {
		return nil, errors.New("unauthorized")
	}
	return s, nil
}

func (u *storeUsecase) UpdateStore(userID, storeID int64, name string) error {
	s, err := u.repo.GetByID(storeID)
	if err != nil {
		return err
	}
	if s.UserID != userID {
		return errors.New("unauthorized")
	}
	return u.repo.Update(storeID, name)
}

func (u *storeUsecase) DeleteStore(userID, storeID int64) error {
	s, err := u.repo.GetByID(storeID)
	if err != nil {
		return err
	}
	if s.UserID != userID {
		return errors.New("unauthorized")
	}
	return u.repo.Delete(storeID)
}

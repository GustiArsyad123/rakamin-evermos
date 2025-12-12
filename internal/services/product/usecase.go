package product

import (
	"errors"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

type Usecase interface {
	CreateProduct(userID int64, p *models.Product) (int64, error)
	ListProducts(filters map[string]string, page, limit int) ([]*models.Product, int, error)
	GetProduct(id int64) (*models.Product, error)
	UpdateStore(userID, storeID int64, name string) error
}

type productUsecase struct {
	repo Repository
}

func NewUsecase(r Repository) Usecase {
	return &productUsecase{repo: r}
}

func (u *productUsecase) CreateProduct(userID int64, p *models.Product) (int64, error) {
	// find store id by user
	var storeID int64
	row := u.repo.(*mysqlRepo).db.QueryRow("SELECT id FROM stores WHERE user_id = ?", userID)
	if err := row.Scan(&storeID); err != nil {
		return 0, err
	}
	p.StoreID = storeID
	return u.repo.Create(p)
}

func (u *productUsecase) ListProducts(filters map[string]string, page, limit int) ([]*models.Product, int, error) {
	return u.repo.List(filters, page, limit)
}

func (u *productUsecase) GetProduct(id int64) (*models.Product, error) {
	return u.repo.GetByID(id)
}

func (u *productUsecase) UpdateStore(userID, storeID int64, name string) error {
	// Check ownership
	store, err := u.repo.GetStoreByID(storeID)
	if err != nil {
		return err
	}
	if store == nil {
		return errors.New("store not found")
	}
	if store.UserID != userID {
		return errors.New("forbidden")
	}
	return u.repo.UpdateStore(storeID, name)
}

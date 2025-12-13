package product

import (
	"errors"
	"strconv"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

type Usecase interface {
	CreateProduct(userID int64, p *models.Product) (int64, error)
	ListProducts(userID int64, filters map[string]string, page, limit int) ([]*models.Product, int, error)
	GetProduct(userID, id int64) (*models.Product, error)
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

func (u *productUsecase) ListProducts(userID int64, filters map[string]string, page, limit int) ([]*models.Product, int, error) {
	// find store id by user
	var storeID int64
	row := u.repo.(*mysqlRepo).db.QueryRow("SELECT id FROM stores WHERE user_id = ?", userID)
	if err := row.Scan(&storeID); err != nil {
		return nil, 0, err
	}
	filters["store_id"] = strconv.FormatInt(storeID, 10)
	return u.repo.List(filters, page, limit)
}

func (u *productUsecase) GetProduct(userID, id int64) (*models.Product, error) {
	p, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	// find store id by user
	var storeID int64
	row := u.repo.(*mysqlRepo).db.QueryRow("SELECT id FROM stores WHERE user_id = ?", userID)
	if err := row.Scan(&storeID); err != nil {
		return nil, err
	}
	if p.StoreID != storeID {
		return nil, errors.New("unauthorized")
	}
	return p, nil
}

package transaction

import (
	"database/sql"
	"errors"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

type Usecase interface {
	Create(userID int64, items []ItemReq) (int64, error)
	Get(userID, id int64) (*models.Transaction, []*models.ProductLog, error)
	List(userID int64, page, limit int) ([]*models.Transaction, int, error)
}

type ItemReq struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type txnUsecase struct {
	repo Repository
	db   *sql.DB
}

func NewUsecase(r Repository, db *sql.DB) Usecase {
	return &txnUsecase{repo: r, db: db}
}

func (u *txnUsecase) Create(userID int64, items []ItemReq) (int64, error) {
	if len(items) == 0 {
		return 0, errors.New("no items")
	}
	// load products and ensure same store
	var storeID int64 = 0
	logs := []*models.ProductLog{}
	total := 0.0
	for _, it := range items {
		p := &models.Product{}
		row := u.db.QueryRow("SELECT id,store_id,name,price,stock FROM products WHERE id = ?", it.ProductID)
		if err := row.Scan(&p.ID, &p.StoreID, &p.Name, &p.Price, &p.Stock); err != nil {
			if err == sql.ErrNoRows {
				return 0, errors.New("product not found")
			}
			return 0, err
		}
		if it.Quantity <= 0 {
			return 0, errors.New("invalid quantity")
		}
		if p.Stock < it.Quantity {
			return 0, errors.New("insufficient stock")
		}
		if storeID == 0 {
			storeID = p.StoreID
		}
		if p.StoreID != storeID {
			return 0, errors.New("items must be from same store")
		}
		total += p.Price * float64(it.Quantity)
		logs = append(logs, &models.ProductLog{ProductID: p.ID, ProductName: p.Name, ProductPrice: p.Price, Quantity: it.Quantity})
	}

	txn := &models.Transaction{UserID: userID, StoreID: storeID, Total: total, Status: "pending"}
	id, err := u.repo.Create(txn, logs)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (u *txnUsecase) Get(userID, id int64) (*models.Transaction, []*models.ProductLog, error) {
	t, logs, err := u.repo.GetByID(id)
	if err != nil || t == nil {
		return t, logs, err
	}
	// Skip ownership check if userID == 0 (for testing without auth)
	if userID != 0 && t.UserID != userID {
		return nil, nil, errors.New("forbidden")
	}
	return t, logs, nil
}

func (u *txnUsecase) List(userID int64, page, limit int) ([]*models.Transaction, int, error) {
	return u.repo.ListByUser(userID, page, limit)
}

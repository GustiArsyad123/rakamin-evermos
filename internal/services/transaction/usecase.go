package transaction

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/example/ms-ecommerce/internal/pkg/models"
	"github.com/example/ms-ecommerce/internal/pkg/payment"
)

type Usecase interface {
	Create(userID int64, addressID int64, items []ItemReq) (int64, error)
	// CreateAndCharge creates a transaction and attempts to charge via the configured payment provider.
	CreateAndCharge(userID int64, addressID int64, items []ItemReq, paymentMethod string, paymentToken string) (int64, error)
	Get(userID, id int64, role string) (*models.Transaction, []*models.ProductLog, error)
	List(userID int64, role string, filters map[string]string, page, limit int) ([]*models.Transaction, int, error)
}

type ItemReq struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type txnUsecase struct {
	repo    Repository
	db      *sql.DB
	payment payment.PaymentProvider
}

func NewUsecase(r Repository, db *sql.DB, p payment.PaymentProvider) Usecase {
	return &txnUsecase{repo: r, db: db, payment: p}
}

// helper to validate items and compute total and logs (does not modify DB)
func (u *txnUsecase) validateAndCompute(items []ItemReq) (int64, []*models.ProductLog, float64, error) {
	if len(items) == 0 {
		return 0, nil, 0, errors.New("no items")
	}
	var storeID int64 = 0
	logs := []*models.ProductLog{}
	total := 0.0
	for _, it := range items {
		p := &models.Product{}
		row := u.db.QueryRow("SELECT id,store_id,name,price,stock FROM products WHERE id = ?", it.ProductID)
		if err := row.Scan(&p.ID, &p.StoreID, &p.Name, &p.Price, &p.Stock); err != nil {
			if err == sql.ErrNoRows {
				return 0, nil, 0, errors.New("product not found")
			}
			return 0, nil, 0, err
		}
		if it.Quantity <= 0 {
			return 0, nil, 0, errors.New("invalid quantity")
		}
		if p.Stock < it.Quantity {
			return 0, nil, 0, errors.New("insufficient stock")
		}
		if storeID == 0 {
			storeID = p.StoreID
		}
		if p.StoreID != storeID {
			return 0, nil, 0, errors.New("items must be from same store")
		}
		total += p.Price * float64(it.Quantity)
		logs = append(logs, &models.ProductLog{ProductID: p.ID, ProductName: p.Name, ProductPrice: p.Price, Quantity: it.Quantity})
	}
	return storeID, logs, total, nil
}

func (u *txnUsecase) Create(userID int64, addressID int64, items []ItemReq) (int64, error) {
	// validate address
	var addrUserID int64
	if err := u.db.QueryRow("SELECT user_id FROM addresses WHERE id = ?", addressID).Scan(&addrUserID); err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("address not found")
		}
		return 0, err
	}
	if addrUserID != userID {
		return 0, errors.New("address does not belong to user")
	}

	storeID, logs, total, err := u.validateAndCompute(items)
	if err != nil {
		return 0, err
	}

	txn := &models.Transaction{UserID: userID, StoreID: storeID, AddressID: addressID, Total: total, Status: "pending"}
	id, err := u.repo.Create(txn, logs)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// CreateAndCharge creates the transaction and attempts to charge using the configured payment provider.
func (u *txnUsecase) CreateAndCharge(userID int64, addressID int64, items []ItemReq, paymentMethod string, paymentToken string) (int64, error) {
	// validate address
	var addrUserID int64
	if err := u.db.QueryRow("SELECT user_id FROM addresses WHERE id = ?", addressID).Scan(&addrUserID); err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("address not found")
		}
		return 0, err
	}
	if addrUserID != userID {
		return 0, errors.New("address does not belong to user")
	}

	storeID, logs, total, err := u.validateAndCompute(items)
	if err != nil {
		return 0, err
	}

	// Create transaction (will decrease stock and set status pending)
	txn := &models.Transaction{UserID: userID, StoreID: storeID, AddressID: addressID, Total: total, Status: "pending"}
	id, err := u.repo.Create(txn, logs)
	if err != nil {
		return 0, err
	}

	// If no payment provider configured, return pending id
	if u.payment == nil {
		return id, nil
	}

	// Attempt to charge
	metadata := map[string]string{"transaction_id": strconv.FormatInt(id, 10), "method": paymentMethod}
	providerTxnID, err := u.payment.Charge(total, "USD", paymentToken, metadata)
	if err != nil {
		// payment failed: attempt compensation to revert stock and mark transaction failed
		// fetch logs to know quantities
		_, logs, getErr := u.repo.GetByID(id)
		if getErr == nil {
			tx, txErr := u.db.Begin()
			if txErr == nil {
				for _, l := range logs {
					_, _ = tx.Exec("UPDATE products SET stock = stock + ? WHERE id = ?", l.Quantity, l.ProductID)
				}
				// persist failure status
				metaBytes, _ := json.Marshal(map[string]string{"error": err.Error()})
				_, _ = tx.Exec("UPDATE transactions SET status = ?, payment_metadata = ? WHERE id = ?", "failed", string(metaBytes), id)
				_ = tx.Commit()
			}
		} else {
			// fallback: mark transaction failed if we couldn't get logs
			_, _ = u.db.Exec("UPDATE transactions SET status = ? WHERE id = ?", "failed", id)
		}
		return id, fmt.Errorf("payment failed: %w", err)
	}

	// payment succeeded: persist provider txn id and metadata and mark paid
	metaBytes, _ := json.Marshal(metadata)
	_, _ = u.db.Exec("UPDATE transactions SET status = ?, provider_txn_id = ?, payment_metadata = ? WHERE id = ?", "paid", providerTxnID, string(metaBytes), id)
	return id, nil
}

func (u *txnUsecase) Get(userID, id int64, role string) (*models.Transaction, []*models.ProductLog, error) {
	t, logs, err := u.repo.GetByID(id)
	if err != nil || t == nil {
		return t, logs, err
	}
	if role != "admin" && t.UserID != userID {
		return nil, nil, errors.New("forbidden")
	}
	return t, logs, nil
}

func (u *txnUsecase) List(userID int64, role string, filters map[string]string, page, limit int) ([]*models.Transaction, int, error) {
	// If admin, list all transactions (userID=0), else list user's transactions
	listUserID := userID
	if role == "admin" {
		listUserID = 0
	}
	return u.repo.ListByUser(listUserID, filters, page, limit)
}

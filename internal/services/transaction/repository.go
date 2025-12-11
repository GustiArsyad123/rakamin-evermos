package transaction

import (
	"database/sql"
	"fmt"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

type Repository interface {
	Create(txn *models.Transaction, logs []*models.ProductLog) (int64, error)
	GetByID(id int64) (*models.Transaction, []*models.ProductLog, error)
	ListByUser(userID int64, page, limit int) ([]*models.Transaction, int, error)
}

type mysqlRepo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repository {
	return &mysqlRepo{db: db}
}

func (r *mysqlRepo) Create(txn *models.Transaction, logs []*models.ProductLog) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	res, err := tx.Exec("INSERT INTO transactions (user_id,store_id,total,status) VALUES (?,?,?,?)", txn.UserID, txn.StoreID, txn.Total, txn.Status)
	if err != nil {
		return 0, err
	}
	tid, _ := res.LastInsertId()

	// insert logs and update stock
	for _, l := range logs {
		_, err = tx.Exec("INSERT INTO product_logs (transaction_id,product_id,product_name,product_price,quantity) VALUES (?,?,?,?,?)", tid, l.ProductID, l.ProductName, l.ProductPrice, l.Quantity)
		if err != nil {
			return 0, err
		}
		// decrease stock if possible
		resu, err := tx.Exec("UPDATE products SET stock = stock - ? WHERE id = ? AND stock >= ?", l.Quantity, l.ProductID, l.Quantity)
		if err != nil {
			return 0, err
		}
		ra, _ := resu.RowsAffected()
		if ra == 0 {
			return 0, fmt.Errorf("insufficient stock for product %d", l.ProductID)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return tid, nil
}

func (r *mysqlRepo) GetByID(id int64) (*models.Transaction, []*models.ProductLog, error) {
	t := &models.Transaction{}
	row := r.db.QueryRow("SELECT id,user_id,store_id,total,status,created_at FROM transactions WHERE id = ?", id)
	if err := row.Scan(&t.ID, &t.UserID, &t.StoreID, &t.Total, &t.Status, &t.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	rows, err := r.db.Query("SELECT id,transaction_id,product_id,product_name,product_price,quantity,created_at FROM product_logs WHERE transaction_id = ?", id)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	logs := []*models.ProductLog{}
	for rows.Next() {
		l := &models.ProductLog{}
		if err := rows.Scan(&l.ID, &l.TransactionID, &l.ProductID, &l.ProductName, &l.ProductPrice, &l.Quantity, &l.CreatedAt); err != nil {
			return nil, nil, err
		}
		logs = append(logs, l)
	}
	return t, logs, nil
}

func (r *mysqlRepo) ListByUser(userID int64, page, limit int) ([]*models.Transaction, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	var total int
	var countQuery string
	var listQuery string
	if userID == 0 {
		// List all transactions for testing
		countQuery = "SELECT COUNT(1) FROM transactions"
		listQuery = "SELECT id,user_id,store_id,total,status,created_at FROM transactions ORDER BY created_at DESC LIMIT ? OFFSET ?"
	} else {
		countQuery = "SELECT COUNT(1) FROM transactions WHERE user_id = ?"
		listQuery = "SELECT id,user_id,store_id,total,status,created_at FROM transactions WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	}
	if userID == 0 {
		if err := r.db.QueryRow(countQuery).Scan(&total); err != nil {
			return nil, 0, err
		}
		rows, err := r.db.Query(listQuery, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		defer rows.Close()
		out := []*models.Transaction{}
		for rows.Next() {
			t := &models.Transaction{}
			if err := rows.Scan(&t.ID, &t.UserID, &t.StoreID, &t.Total, &t.Status, &t.CreatedAt); err != nil {
				return nil, 0, err
			}
			out = append(out, t)
		}
		return out, total, nil
	} else {
		if err := r.db.QueryRow(countQuery, userID).Scan(&total); err != nil {
			return nil, 0, err
		}
		rows, err := r.db.Query(listQuery, userID, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		defer rows.Close()
		out := []*models.Transaction{}
		for rows.Next() {
			t := &models.Transaction{}
			if err := rows.Scan(&t.ID, &t.UserID, &t.StoreID, &t.Total, &t.Status, &t.CreatedAt); err != nil {
				return nil, 0, err
			}
			out = append(out, t)
		}
		return out, total, nil
	}
}

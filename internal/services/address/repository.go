package address

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

type Repository interface {
	Create(a *models.Address) (int64, error)
	ListByUser(userID int64, filters map[string]string, page, limit int) ([]*models.Address, int, error)
	GetByID(id int64) (*models.Address, error)
	Update(id int64, label, address, city, postalCode string) error
	Delete(id int64) error
}

type mysqlRepo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repository {
	return &mysqlRepo{db: db}
}

func (r *mysqlRepo) Create(a *models.Address) (int64, error) {
	res, err := r.db.Exec("INSERT INTO addresses (user_id,label,address,city,postal_code) VALUES (?,?,?,?,?)",
		a.UserID, a.Label, a.Address, a.City, a.PostalCode)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

func (r *mysqlRepo) ListByUser(userID int64, filters map[string]string, page, limit int) ([]*models.Address, int, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	if userID != 0 {
		where = append(where, "user_id = ?")
		args = append(args, userID)
	}
	if v, ok := filters["label"]; ok && v != "" {
		where = append(where, "label LIKE ?")
		args = append(args, "%"+v+"%")
	}
	if v, ok := filters["city"]; ok && v != "" {
		where = append(where, "city = ?")
		args = append(args, v)
	}
	if v, ok := filters["postal_code"]; ok && v != "" {
		where = append(where, "postal_code = ?")
		args = append(args, v)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM addresses WHERE %s", strings.Join(where, " AND "))
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	q := fmt.Sprintf("SELECT id,user_id,label,address,city,postal_code,created_at FROM addresses WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?", strings.Join(where, " AND "))
	args = append(args, limit, offset)
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []*models.Address{}
	for rows.Next() {
		a := &models.Address{}
		if err := rows.Scan(&a.ID, &a.UserID, &a.Label, &a.Address, &a.City, &a.PostalCode, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, a)
	}
	return out, total, nil
}

func (r *mysqlRepo) GetByID(id int64) (*models.Address, error) {
	a := &models.Address{}
	row := r.db.QueryRow("SELECT id,user_id,label,address,city,postal_code,created_at FROM addresses WHERE id = ?", id)
	err := row.Scan(&a.ID, &a.UserID, &a.Label, &a.Address, &a.City, &a.PostalCode, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return a, err
}

func (r *mysqlRepo) Update(id int64, label, address, city, postalCode string) error {
	_, err := r.db.Exec("UPDATE addresses SET label=?, address=?, city=?, postal_code=? WHERE id=?", label, address, city, postalCode, id)
	return err
}

func (r *mysqlRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM addresses WHERE id=?", id)
	return err
}

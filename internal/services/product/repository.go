package product

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

type Repository interface {
	Create(p *models.Product) (int64, error)
	List(filters map[string]string, page, limit int) ([]*models.Product, int, error)
	GetByID(id int64) (*models.Product, error)
	GetStoreByID(id int64) (*models.Store, error)
	UpdateStore(id int64, name string) error
}

type mysqlRepo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repository {
	return &mysqlRepo{db: db}
}

func (r *mysqlRepo) Create(p *models.Product) (int64, error) {
	// if a category id is provided, ensure it exists to avoid FK errors
	if p.CategoryID != nil {
		var exists int
		err := r.db.QueryRow("SELECT 1 FROM categories WHERE id = ?", *p.CategoryID).Scan(&exists)
		if err != nil {
			if err == sql.ErrNoRows {
				return 0, fmt.Errorf("category %d not found", *p.CategoryID)
			}
			return 0, err
		}
	}
	res, err := r.db.Exec("INSERT INTO products (store_id,category_id,name,description,price,stock,image_url) VALUES (?,?,?,?,?,?,?)",
		p.StoreID, p.CategoryID, p.Name, p.Description, p.Price, p.Stock, p.ImageURL)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

func (r *mysqlRepo) GetByID(id int64) (*models.Product, error) {
	p := &models.Product{}
	row := r.db.QueryRow("SELECT id,store_id,category_id,name,description,price,stock,image_url,created_at FROM products WHERE id = ?", id)
	var cat sql.NullInt64
	err := row.Scan(&p.ID, &p.StoreID, &cat, &p.Name, &p.Description, &p.Price, &p.Stock, &p.ImageURL, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if cat.Valid {
		v := cat.Int64
		p.CategoryID = &v
	}
	return p, nil
}

func (r *mysqlRepo) List(filters map[string]string, page, limit int) ([]*models.Product, int, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	if v, ok := filters["search"]; ok && v != "" {
		where = append(where, "(name LIKE ? OR description LIKE ?)")
		like := "%" + v + "%"
		args = append(args, like, like)
	}
	if v, ok := filters["category_id"]; ok && v != "" {
		where = append(where, "category_id = ?")
		args = append(args, v)
	}
	if v, ok := filters["store_id"]; ok && v != "" {
		where = append(where, "store_id = ?")
		args = append(args, v)
	}
	// price filters
	if v, ok := filters["min_price"]; ok && v != "" {
		where = append(where, "price >= ?")
		args = append(args, v)
	}
	if v, ok := filters["max_price"]; ok && v != "" {
		where = append(where, "price <= ?")
		args = append(args, v)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM products WHERE %s", strings.Join(where, " AND "))
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

	q := fmt.Sprintf("SELECT id,store_id,category_id,name,description,price,stock,image_url,created_at FROM products WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?", strings.Join(where, " AND "))
	args = append(args, limit, offset)
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []*models.Product{}
	for rows.Next() {
		p := &models.Product{}
		var cat sql.NullInt64
		if err := rows.Scan(&p.ID, &p.StoreID, &cat, &p.Name, &p.Description, &p.Price, &p.Stock, &p.ImageURL, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		if cat.Valid {
			v := cat.Int64
			p.CategoryID = &v
		}
		out = append(out, p)
	}
	return out, total, nil
}

func (r *mysqlRepo) GetStoreByID(id int64) (*models.Store, error) {
	s := &models.Store{}
	row := r.db.QueryRow("SELECT id,user_id,name,created_at FROM stores WHERE id = ?", id)
	err := row.Scan(&s.ID, &s.UserID, &s.Name, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *mysqlRepo) UpdateStore(id int64, name string) error {
	_, err := r.db.Exec("UPDATE stores SET name = ? WHERE id = ?", name, id)
	return err
}

package category

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

type Repository interface {
	Create(name string) (int64, error)
	Update(id int64, name string) error
	Delete(id int64) error
	GetByID(id int64) (*models.Category, error)
	List(filters map[string]string, page, limit int) ([]*models.Category, int, error)
}

type mysqlRepo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repository {
	return &mysqlRepo{db: db}
}

func (r *mysqlRepo) Create(name string) (int64, error) {
	res, err := r.db.Exec("INSERT INTO categories (name) VALUES (?)", name)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

func (r *mysqlRepo) Update(id int64, name string) error {
	_, err := r.db.Exec("UPDATE categories SET name = ? WHERE id = ?", name, id)
	return err
}

func (r *mysqlRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM categories WHERE id = ?", id)
	return err
}

func (r *mysqlRepo) GetByID(id int64) (*models.Category, error) {
	c := &models.Category{}
	row := r.db.QueryRow("SELECT id,name,created_at FROM categories WHERE id = ?", id)
	if err := row.Scan(&c.ID, &c.Name, &c.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return c, nil
}

func (r *mysqlRepo) List(filters map[string]string, page, limit int) ([]*models.Category, int, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	if v, ok := filters["search"]; ok && v != "" {
		where = append(where, "name LIKE ?")
		args = append(args, "%"+v+"%")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM categories WHERE %s", strings.Join(where, " AND "))
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

	listQuery := fmt.Sprintf("SELECT id,name,created_at FROM categories WHERE %s ORDER BY name ASC LIMIT ? OFFSET ?", strings.Join(where, " AND "))
	args = append(args, limit, offset)

	rows, err := r.db.Query(listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []*models.Category{}
	for rows.Next() {
		c := &models.Category{}
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, total, nil
}

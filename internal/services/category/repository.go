package category

import (
	"database/sql"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

type Repository interface {
	Create(name string) (int64, error)
	Update(id int64, name string) error
	Delete(id int64) error
	GetByID(id int64) (*models.Category, error)
	List() ([]*models.Category, error)
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

func (r *mysqlRepo) List() ([]*models.Category, error) {
	rows, err := r.db.Query("SELECT id,name,created_at FROM categories ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*models.Category{}
	for rows.Next() {
		c := &models.Category{}
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

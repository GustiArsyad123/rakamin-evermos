package store

import (
	"database/sql"

	"github.com/example/ms-ecommerce/internal/pkg/models"
)

type Repository interface {
	Create(s *models.Store) (int64, error)
	GetByID(id int64) (*models.Store, error)
	GetByUserID(userID int64) (*models.Store, error)
	Update(id int64, name string) error
	Delete(id int64) error
}

type mysqlRepo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repository {
	return &mysqlRepo{db: db}
}

func (r *mysqlRepo) Create(s *models.Store) (int64, error) {
	res, err := r.db.Exec("INSERT INTO stores (user_id, name) VALUES (?, ?)", s.UserID, s.Name)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

func (r *mysqlRepo) GetByID(id int64) (*models.Store, error) {
	s := &models.Store{}
	row := r.db.QueryRow("SELECT id, user_id, name, created_at FROM stores WHERE id = ?", id)
	err := row.Scan(&s.ID, &s.UserID, &s.Name, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *mysqlRepo) GetByUserID(userID int64) (*models.Store, error) {
	s := &models.Store{}
	row := r.db.QueryRow("SELECT id, user_id, name, created_at FROM stores WHERE user_id = ?", userID)
	err := row.Scan(&s.ID, &s.UserID, &s.Name, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *mysqlRepo) Update(id int64, name string) error {
	_, err := r.db.Exec("UPDATE stores SET name = ? WHERE id = ?", name, id)
	return err
}

func (r *mysqlRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM stores WHERE id = ?", id)
	return err
}
